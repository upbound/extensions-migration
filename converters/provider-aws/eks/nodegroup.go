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

package eks

import (
	srcv1alpha1 "github.com/crossplane-contrib/provider-aws/apis/eks/manualv1alpha1"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/crossplane/upjet/pkg/migration"
	"github.com/pkg/errors"
	"github.com/upbound/extensions-migration/converters/common"
	targetv1beta1 "github.com/upbound/provider-aws/apis/eks/v1beta1"
)

func NodegroupResource(mg resource.Managed) ([]resource.Managed, error) {
	source := mg.(*srcv1alpha1.NodeGroup)
	target := &targetv1beta1.NodeGroup{}
	if _, err := migration.CopyInto(source, target, targetv1beta1.NodeGroup_GroupVersionKind, "spec.forProvider.scalingConfig"); err != nil {
		return nil, errors.Wrap(err, "failed to copy source into target")
	}

	scalingConfig := &targetv1beta1.ScalingConfigParameters{
		DesiredSize: common.PtrFloat64FromInt32(source.Spec.ForProvider.ScalingConfig.DesiredSize),
		MinSize:     common.PtrFloat64FromInt32(source.Spec.ForProvider.ScalingConfig.MinSize),
		MaxSize:     common.PtrFloat64FromInt32(source.Spec.ForProvider.ScalingConfig.MaxSize),
	}

	target.Spec.ForProvider.ScalingConfig = append(target.Spec.ForProvider.ScalingConfig, *scalingConfig)

	if len(source.Spec.ForProvider.NodeRole) > 0 {
		target.Spec.ForProvider.NodeRoleArn = &source.Spec.ForProvider.NodeRole
	}

	target.Spec.ForProvider.NodeRoleArnSelector = source.Spec.ForProvider.NodeRoleSelector
	target.Spec.ForProvider.NodeRoleArnRef = source.Spec.ForProvider.NodeRoleRef

	for _, subnet := range source.Spec.ForProvider.Subnets {
		target.Spec.ForProvider.SubnetIds = append(target.Spec.ForProvider.SubnetIds, &subnet)
	}

	target.Spec.ForProvider.SubnetIDRefs = source.Spec.ForProvider.SubnetRefs
	target.Spec.ForProvider.SubnetIDSelector = source.Spec.ForProvider.SubnetSelector

	for _, taints := range source.Spec.ForProvider.Taints {
		taint := &targetv1beta1.TaintParameters{
			Effect: common.PtrFromString(taints.Effect),
			Key:    taints.Key,
			Value:  taints.Value,
		}
		target.Spec.ForProvider.Taint = append(target.Spec.ForProvider.Taint, *taint)
	}

	return []resource.Managed{
		target,
	}, nil
}
