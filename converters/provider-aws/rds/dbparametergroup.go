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
	srcv1alpha1 "github.com/crossplane-contrib/provider-aws/apis/rds/v1alpha1"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/pkg/errors"
	targetv1beta1 "github.com/upbound/provider-aws/apis/rds/v1beta1"
	"github.com/upbound/upjet/pkg/migration"
)

func ParameterGroupResource(mg resource.Managed) ([]resource.Managed, error) {
	source := mg.(*srcv1alpha1.DBParameterGroup)
	target := &targetv1beta1.ParameterGroup{}
	if _, err := migration.CopyInto(source, target, targetv1beta1.ParameterGroup_GroupVersionKind, "spec.forProvider.tags"); err != nil {
		return nil, errors.Wrap(err, "failed to copy source into target")
	}
	target.Spec.ForProvider.Tags = make(map[string]*string, len(source.Spec.ForProvider.Tags))
	for _, t := range source.Spec.ForProvider.Tags {
		v := t.Value
		target.Spec.ForProvider.Tags[*t.Key] = v
	}
	for _, t := range source.Spec.ForProvider.Parameters {
		parameter := &targetv1beta1.ParameterParameters{ApplyMethod: t.ApplyMethod, Name: t.ParameterName, Value: t.ParameterValue}
		target.Spec.ForProvider.Parameter = append(target.Spec.ForProvider.Parameter, *parameter)
	}
	target.Spec.ForProvider.Family = source.Spec.ForProvider.DBParameterGroupFamily

	return []resource.Managed{
		target,
	}, nil
}
