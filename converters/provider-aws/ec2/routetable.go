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
	srcv1alpha1 "github.com/crossplane-contrib/provider-aws/apis/ec2/v1beta1"
	runtimev1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/crossplane/upjet/pkg/migration"
	"github.com/pkg/errors"
	targetv1beta1 "github.com/upbound/provider-aws/apis/ec2/v1beta1"
)

func RouteTableResource(mg resource.Managed) ([]resource.Managed, error) {
	source := mg.(*srcv1alpha1.RouteTable)
	target := &targetv1beta1.RouteTable{}
	if _, err := migration.CopyInto(source, target, targetv1beta1.RouteTable_GroupVersionKind, "spec.forProvider.tags"); err != nil {
		return nil, errors.Wrap(err, "failed to copy source into target")
	}
	target.Spec.ForProvider.Tags = make(map[string]*string, len(source.Spec.ForProvider.Tags))
	for _, t := range source.Spec.ForProvider.Tags {
		v := t.Value
		target.Spec.ForProvider.Tags[t.Key] = &v
	}
	routeTableMRs := []resource.Managed{target}
	externalNameRouteTableId := source.Annotations["crossplane.io/external-name"]
	NameRouteTable := source.Name
	if source.Spec.ForProvider.Associations != nil {
		for _, association := range source.Spec.ForProvider.Associations {
			rtAssociation := &targetv1beta1.RouteTableAssociation{}
			rtAssociation.SetGroupVersionKind(targetv1beta1.RouteTableAssociation_GroupVersionKind)
			rtAssociation.Labels = source.Labels
			rtAssociation.Labels["resourceType"] = "RouteTableAssociation"
			rtAssociation.Spec.DeletionPolicy = source.Spec.DeletionPolicy
			rtAssociation.Spec.ForProvider.Region = &source.Spec.ForProvider.Region
			rtAssociation.Spec.ForProvider.SubnetID = association.SubnetID
			rtAssociation.Spec.ForProvider.SubnetIDSelector = association.SubnetIDSelector
			rtAssociation.Spec.ForProvider.SubnetIDRef = association.SubnetIDRef
			matchController := true
			if len(externalNameRouteTableId) > 0 {
				rtAssociation.Spec.ForProvider.RouteTableID = &externalNameRouteTableId
			}
			rtAssociation.Spec.ForProvider.RouteTableIDRef = &runtimev1.Reference{}
			rtAssociation.Spec.ForProvider.RouteTableIDRef.Name = NameRouteTable
			rtAssociation.Spec.ForProvider.RouteTableIDSelector = &runtimev1.Selector{}
			rtAssociation.Spec.ForProvider.RouteTableIDSelector.MatchLabels = source.Labels
			rtAssociation.Spec.ForProvider.RouteTableIDSelector.MatchControllerRef = &matchController
			rtAssociation.Spec.ProviderConfigReference = source.Spec.ProviderConfigReference
			routeTableMRs = append(routeTableMRs, rtAssociation)
		}
	}
	if source.Spec.ForProvider.Routes != nil {
		for _, route := range source.Spec.ForProvider.Routes {
			rtRoute := &targetv1beta1.Route{}
			rtRoute.SetGroupVersionKind(targetv1beta1.Route_GroupVersionKind)
			rtRoute.Labels = source.Labels
			rtRoute.Labels["resourceType"] = "Route"
			rtRoute.Spec.DeletionPolicy = source.Spec.DeletionPolicy
			rtRoute.Spec.ForProvider.Region = &source.Spec.ForProvider.Region
			rtRoute.Spec.ForProvider.NATGatewayID = route.NatGatewayID
			rtRoute.Spec.ForProvider.NATGatewayIDSelector = route.NatGatewayIDSelector
			rtRoute.Spec.ForProvider.NATGatewayIDRef = route.NatGatewayIDRef
			rtRoute.Spec.ForProvider.DestinationCidrBlock = route.DestinationCIDRBlock
			rtRoute.Spec.ForProvider.TransitGatewayID = route.TransitGatewayID
			matchController := true
			if len(externalNameRouteTableId) > 0 {
				rtRoute.Spec.ForProvider.RouteTableID = &externalNameRouteTableId
			}
			rtRoute.Spec.ForProvider.RouteTableIDRef = &runtimev1.Reference{}
			rtRoute.Spec.ForProvider.RouteTableIDRef.Name = NameRouteTable
			rtRoute.Spec.ForProvider.RouteTableIDSelector = &runtimev1.Selector{}
			rtRoute.Spec.ForProvider.RouteTableIDSelector.MatchLabels = source.Labels
			rtRoute.Spec.ForProvider.RouteTableIDSelector.MatchControllerRef = &matchController
			rtRoute.Spec.ProviderConfigReference = source.Spec.ProviderConfigReference
			routeTableMRs = append(routeTableMRs, rtRoute)
		}

	}
	return routeTableMRs, nil
}
