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
	runtimev1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/pkg/errors"
	targetv1beta1 "github.com/upbound/provider-aws/apis/ec2/v1beta1"
	"github.com/upbound/upjet/pkg/migration"
)

func VPCEndpointResource(mg resource.Managed) ([]resource.Managed, error) {
	source := mg.(*srcv1alpha1.VPCEndpoint)
	target := &targetv1beta1.VPCEndpoint{}
	if _, err := migration.CopyInto(source, target, targetv1beta1.VPCEndpoint_GroupVersionKind); err != nil {
		return nil, errors.Wrap(err, "failed to copy source into target")
	}
	target.Spec.ForProvider.PrivateDNSEnabled = source.Spec.ForProvider.PrivateDNSEnabled

	vpcEndpointMRs := []resource.Managed{target}
	externalNameVpcEndpointId := source.Annotations["crossplane.io/external-name"]
	NameVpcEndpoint := source.Name

	// for offline resources
	if source.Spec.ForProvider.SubnetIDSelector != nil && len(source.Spec.ForProvider.SubnetIDs) <= 1 {
		vpcEndpointSubnetAssociation := &targetv1beta1.VPCEndpointSubnetAssociation{}
		vpcEndpointSubnetAssociation.SetGroupVersionKind(targetv1beta1.VPCEndpointSubnetAssociation_GroupVersionKind)
		vpcEndpointSubnetAssociation.Labels = source.Labels
		vpcEndpointSubnetAssociation.Labels["resourceType"] = "VPCEndpointSubnetAssociation"
		vpcEndpointSubnetAssociation.Spec.DeletionPolicy = source.Spec.DeletionPolicy
		vpcEndpointSubnetAssociation.Spec.ForProvider.Region = &source.Spec.ForProvider.Region
		vpcEndpointSubnetAssociation.Spec.ForProvider.SubnetIDSelector = source.Spec.ForProvider.SubnetIDSelector
		if len(externalNameVpcEndpointId) > 0 {
			vpcEndpointSubnetAssociation.Spec.ForProvider.VPCEndpointID = &externalNameVpcEndpointId
		}
		vpcEndpointSubnetAssociation.Spec.ForProvider.VPCEndpointIDRef = &runtimev1.Reference{}
		vpcEndpointSubnetAssociation.Spec.ForProvider.VPCEndpointIDRef.Name = NameVpcEndpoint
		vpcEndpointSubnetAssociation.Spec.ForProvider.VPCEndpointIDSelector = &runtimev1.Selector{}
		vpcEndpointSubnetAssociation.Spec.ForProvider.VPCEndpointIDSelector.MatchLabels = source.Labels
		matchController := true
		vpcEndpointSubnetAssociation.Spec.ForProvider.VPCEndpointIDSelector.MatchControllerRef = &matchController
		vpcEndpointSubnetAssociation.Spec.ProviderConfigReference = source.Spec.ProviderConfigReference
		vpcEndpointMRs = append(vpcEndpointMRs, vpcEndpointSubnetAssociation)
	}

	if len(source.Spec.ForProvider.SubnetIDs) > 1 {
		for _, subnets := range source.Spec.ForProvider.SubnetIDs {
			vpcEndpointSubnetAssociation := &targetv1beta1.VPCEndpointSubnetAssociation{}
			vpcEndpointSubnetAssociation.SetGroupVersionKind(targetv1beta1.VPCEndpointSubnetAssociation_GroupVersionKind)
			vpcEndpointSubnetAssociation.Labels = source.Labels
			vpcEndpointSubnetAssociation.Labels["resourceType"] = "VPCEndpointSubnetAssociation"
			vpcEndpointSubnetAssociation.Spec.DeletionPolicy = source.Spec.DeletionPolicy
			vpcEndpointSubnetAssociation.Spec.ForProvider.Region = &source.Spec.ForProvider.Region
			vpcEndpointSubnetAssociation.Spec.ForProvider.SubnetID = subnets
			if len(externalNameVpcEndpointId) > 0 {
				vpcEndpointSubnetAssociation.Spec.ForProvider.VPCEndpointID = &externalNameVpcEndpointId
			}
			vpcEndpointSubnetAssociation.Spec.ForProvider.VPCEndpointIDRef = &runtimev1.Reference{}
			vpcEndpointSubnetAssociation.Spec.ForProvider.VPCEndpointIDRef.Name = NameVpcEndpoint
			vpcEndpointSubnetAssociation.Spec.ForProvider.VPCEndpointIDSelector = &runtimev1.Selector{}
			vpcEndpointSubnetAssociation.Spec.ForProvider.VPCEndpointIDSelector.MatchLabels = source.Labels
			matchController := true
			vpcEndpointSubnetAssociation.Spec.ForProvider.VPCEndpointIDSelector.MatchControllerRef = &matchController
			vpcEndpointSubnetAssociation.Spec.ProviderConfigReference = source.Spec.ProviderConfigReference
			vpcEndpointSubnetAssociation.Spec.ProviderConfigReference = source.Spec.ProviderConfigReference
			vpcEndpointMRs = append(vpcEndpointMRs, vpcEndpointSubnetAssociation)
		}
	}

	return vpcEndpointMRs, nil
}
