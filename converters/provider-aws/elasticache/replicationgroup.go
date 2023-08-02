// Copyright 2022 Upbound Inc.
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

package elasticache

import (
	srcv1beta1 "github.com/crossplane-contrib/provider-aws/apis/cache/v1beta1"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	v1 "github.com/crossplane/crossplane/apis/apiextensions/v1"
	"github.com/pkg/errors"
	"github.com/upbound/extensions-migration/converters/common"
	targetv1beta1 "github.com/upbound/provider-aws/apis/elasticache/v1beta1"
	"github.com/upbound/upjet/pkg/migration"
)

func ReplicationGroupResource(mg resource.Managed) ([]resource.Managed, error) {
	source := mg.(*srcv1beta1.ReplicationGroup)
	target := &targetv1beta1.ReplicationGroup{}
	if _, err := migration.CopyInto(source, target, targetv1beta1.ReplicationGroup_GroupVersionKind, "spec.forProvider.tags"); err != nil {
		return nil, errors.Wrap(err, "failed to copy source into target")
	}
	target.Spec.ForProvider.Tags = make(map[string]*string, len(source.Spec.ForProvider.Tags))
	for _, t := range source.Spec.ForProvider.Tags {
		v := t.Value
		target.Spec.ForProvider.Tags[t.Key] = &v
	}
	target.Spec.ForProvider.ParameterGroupName = source.Spec.ForProvider.CacheParameterGroupName
	target.Spec.ForProvider.ApplyImmediately = &source.Spec.ForProvider.ApplyModificationsImmediately
	target.Spec.ForProvider.MultiAzEnabled = source.Spec.ForProvider.MultiAZEnabled
	target.Spec.ForProvider.SubnetGroupName = source.Spec.ForProvider.CacheSubnetGroupName
	target.Spec.ForProvider.NodeType = &source.Spec.ForProvider.CacheNodeType
	target.Spec.ForProvider.MaintenanceWindow = source.Spec.ForProvider.PreferredMaintenanceWindow

	return []resource.Managed{
		target,
	}, nil
}

func ReplicationGroupComposition(sourceTemplate v1.ComposedTemplate, convertedTemplates ...*v1.ComposedTemplate) error {
	conversionMap := map[string]string{
		"spec.forProvider.cacheParameterGroupName": "spec.forProvider.parameterGroupName",
	}
	return common.DefaultCompositionConverter(true, conversionMap, sourceTemplate, convertedTemplates...)
}
