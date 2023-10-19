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

package rds

import (
	srcv1beta1 "github.com/crossplane-contrib/provider-aws/apis/database/v1beta1"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/crossplane/upjet/pkg/migration"
	"github.com/pkg/errors"
	targetv1beta1 "github.com/upbound/provider-aws/apis/rds/v1beta1"
)

func InstanceResource(mg resource.Managed) ([]resource.Managed, error) {
	source := mg.(*srcv1beta1.RDSInstance)
	target := &targetv1beta1.Instance{}
	if _, err := migration.CopyInto(source, target, targetv1beta1.Instance_GroupVersionKind, "spec.forProvider.tags"); err != nil {
		return nil, errors.Wrap(err, "failed to copy source into target")
	}
	target.Spec.ForProvider.Tags = make(map[string]*string, len(source.Spec.ForProvider.Tags))
	for _, t := range source.Spec.ForProvider.Tags {
		v := t.Value
		target.Spec.ForProvider.Tags[t.Key] = &v
	}

	target.Spec.ForProvider.ApplyImmediately = source.Spec.ForProvider.ApplyModificationsImmediately
	target.Spec.ForProvider.CACertIdentifier = source.Spec.ForProvider.CACertificateIdentifier
	target.Spec.ForProvider.IAMDatabaseAuthenticationEnabled = source.Spec.ForProvider.EnableIAMDatabaseAuthentication
	target.Spec.ForProvider.PerformanceInsightsEnabled = source.Spec.ForProvider.EnablePerformanceInsights
	target.Spec.ForProvider.FinalSnapshotIdentifier = source.Spec.ForProvider.FinalDBSnapshotIdentifier
	target.Spec.ForProvider.Username = source.Spec.ForProvider.MasterUsername

	return []resource.Managed{
		target,
	}, nil
}
