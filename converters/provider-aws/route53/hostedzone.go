// Copyright 2023 Upbound Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package route53

import (
	srcv1alpha1 "github.com/crossplane-contrib/provider-aws/apis/route53/v1alpha1"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/crossplane/upjet/pkg/migration"
	"github.com/pkg/errors"
	targetv1beta1 "github.com/upbound/provider-aws/apis/route53/v1beta1"
)

func HostedZoneResource(mg resource.Managed) ([]resource.Managed, error) {
	source := mg.(*srcv1alpha1.HostedZone)
	target := &targetv1beta1.Zone{}
	if _, err := migration.CopyInto(source, target, targetv1beta1.Zone_GroupVersionKind, "spec.forProvider.tags", "spec.forProvider.vpc"); err != nil {
		return nil, errors.Wrap(err, "failed to copy source into target")
	}
	// route53 zone is global resource
	region := "eu-central-1"
	target.Spec.ForProvider.Region = &region

	target.Spec.ForProvider.VPC = []targetv1beta1.VPCParameters{
		{
			VPCID:         source.Spec.ForProvider.VPC.VPCID,
			VPCIDSelector: source.Spec.ForProvider.VPC.VPCIDSelector,
			VPCIDRef:      source.Spec.ForProvider.VPC.VPCIDRef,
		},
	}
	return []resource.Managed{
		target,
	}, nil
}
