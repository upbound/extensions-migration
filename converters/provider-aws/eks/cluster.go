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
	srcv1beta1 "github.com/crossplane-contrib/provider-aws/apis/eks/v1beta1"
	runtimev1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/pkg/errors"
	targetv1beta1 "github.com/upbound/provider-aws/apis/eks/v1beta1"
	"github.com/upbound/upjet/pkg/migration"
)

func ClusterResource(mg resource.Managed) ([]resource.Managed, error) {
	source := mg.(*srcv1beta1.Cluster)
	target := &targetv1beta1.Cluster{}
	if _, err := migration.CopyInto(source, target, targetv1beta1.Cluster_GroupVersionKind); err != nil {
		return nil, errors.Wrap(err, "failed to copy source into target")
	}

	eksMRs := []resource.Managed{target}

	vpcConfig := targetv1beta1.VPCConfigParameters{
		EndpointPrivateAccess: source.Spec.ForProvider.ResourcesVpcConfig.EndpointPrivateAccess,
		EndpointPublicAccess:  source.Spec.ForProvider.ResourcesVpcConfig.EndpointPublicAccess,
		SubnetIds:             make([]*string, 0),
		SubnetIDRefs:          append([]runtimev1.Reference{}, source.Spec.ForProvider.ResourcesVpcConfig.SubnetIDRefs...),
		SubnetIDSelector:      source.Spec.ForProvider.ResourcesVpcConfig.SubnetIDSelector,
	}

	for _, subnet := range source.Spec.ForProvider.ResourcesVpcConfig.SubnetIDs {
		vpcConfig.SubnetIds = append(vpcConfig.SubnetIds, &subnet)
	}

	target.Spec.ForProvider.VPCConfig = []targetv1beta1.VPCConfigParameters{vpcConfig}
	// unset because official provider uses ClusterAuth resource for connectionSecret
	target.Spec.WriteConnectionSecretToReference = nil

	if source.Spec.WriteConnectionSecretToReference != nil {
		clusterAuth := &targetv1beta1.ClusterAuth{}
		clusterAuth.SetGroupVersionKind(targetv1beta1.ClusterAuth_GroupVersionKind)
		clusterAuth.Labels = source.Labels
		clusterAuth.Labels["resourceType"] = "ClusterAuth"
		clusterAuth.Spec.DeletionPolicy = source.Spec.DeletionPolicy
		clusterAuth.Spec.ForProvider.Region = *source.Spec.ForProvider.Region
		clusterAuth.Spec.ForProvider.ClusterNameSelector = &runtimev1.Selector{}
		clusterAuth.Spec.ForProvider.ClusterNameSelector.MatchLabels = source.Labels
		matchController := true
		clusterAuth.Spec.ForProvider.ClusterNameSelector.MatchControllerRef = &matchController
		clusterAuth.Spec.ProviderConfigReference = source.Spec.ProviderConfigReference
		clusterAuth.Spec.WriteConnectionSecretToReference = source.Spec.WriteConnectionSecretToReference
		eksMRs = append(eksMRs, clusterAuth)
	}

	return eksMRs, nil
}
