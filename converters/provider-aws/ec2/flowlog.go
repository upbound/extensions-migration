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

package ec2

import (
	srcv1alpha1 "github.com/crossplane-contrib/provider-aws/apis/ec2/v1alpha1"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/crossplane/upjet/pkg/migration"
	"github.com/pkg/errors"
	targetv1beta1 "github.com/upbound/provider-aws/apis/ec2/v1beta1"
)

func FlowLogResource(mg resource.Managed) ([]resource.Managed, error) {
	source := mg.(*srcv1alpha1.FlowLog)
	target := &targetv1beta1.FlowLog{}
	if _, err := migration.CopyInto(source, target, targetv1beta1.FlowLog_GroupVersionKind, "spec.forProvider.tags"); err != nil {
		return nil, errors.Wrap(err, "failed to copy source into target")
	}
	target.Spec.ForProvider.Tags = make(map[string]*string, len(source.Spec.ForProvider.Tags))
	for _, t := range source.Spec.ForProvider.Tags {
		v := t.Value
		if t.Key != nil {
			target.Spec.ForProvider.Tags[*t.Key] = v
		}
	}
	target.Spec.ForProvider.LogDestination = source.Spec.ForProvider.CloudWatchLogDestination
	target.Spec.ForProvider.LogDestinationRef = source.Spec.ForProvider.CloudWatchLogDestinationRef
	target.Spec.ForProvider.LogDestinationSelector = source.Spec.ForProvider.CloudWatchLogDestinationSelector
	target.Spec.ForProvider.IAMRoleArn = source.Spec.ForProvider.DeliverLogsPermissionARN
	target.Spec.ForProvider.IAMRoleArnRef = source.Spec.ForProvider.DeliverLogsPermissionARNRef
	target.Spec.ForProvider.IAMRoleArnSelector = source.Spec.ForProvider.DeliverLogsPermissionARNSelector
	return []resource.Managed{
		target,
	}, nil
}
