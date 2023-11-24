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
	srcv1beta1 "github.com/crossplane-contrib/provider-aws/apis/ec2/v1beta1"
	runtimev1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	v1 "github.com/crossplane/crossplane/apis/apiextensions/v1"
	"github.com/crossplane/upjet/pkg/migration"
	"github.com/pkg/errors"
	targetv1beta1 "github.com/upbound/provider-aws/apis/ec2/v1beta1"

	"github.com/upbound/extensions-migration/converters/common"
	providerawscommon "github.com/upbound/extensions-migration/converters/provider-aws/common"
)

func SecurityGroupResource(mg resource.Managed) ([]resource.Managed, error) {
	source := mg.(*srcv1beta1.SecurityGroup)
	target := &targetv1beta1.SecurityGroup{}
	if _, err := migration.CopyInto(source, target, targetv1beta1.SecurityGroup_GroupVersionKind, "spec.forProvider.tags", "spec.forProvider.ingress", "spec.forProvider.egress"); err != nil {
		return nil, errors.Wrap(err, "failed to copy source into target")
	}
	target.Spec.ForProvider.Tags = make(map[string]*string, len(source.Spec.ForProvider.Tags))
	for _, t := range source.Spec.ForProvider.Tags {
		v := t.Value
		target.Spec.ForProvider.Tags[t.Key] = &v
	}
	target.Spec.ForProvider.Name = &source.Spec.ForProvider.GroupName

	externalNameSecurityGroupId := source.Annotations["crossplane.io/external-name"]
	nameSecurityGroup := source.Name

	secGroupMRs := []resource.Managed{target}
	if source.Spec.ForProvider.Ingress != nil {
		for _, rule := range source.Spec.ForProvider.Ingress {
			sgRule := &targetv1beta1.SecurityGroupRule{}
			sgRule.SetGroupVersionKind(targetv1beta1.SecurityGroupRule_GroupVersionKind)
			sgRule.Spec.ForProvider.Type = common.PtrFromString("ingress")
			sgRule.Spec.ForProvider.Region = source.Spec.ForProvider.Region
			sgRule.Labels = source.Labels
			if len(source.Labels) > 0 {
				sgRule.Labels["resourceType"] = "SecurityGroupRule"
			}
			sgRule.Spec.ForProvider.Protocol = &rule.IPProtocol
			if rule.FromPort != nil {
				convert := float64(*rule.FromPort)
				sgRule.Spec.ForProvider.FromPort = &convert
			}
			if rule.ToPort != nil {
				convert := float64(*rule.ToPort)
				sgRule.Spec.ForProvider.ToPort = &convert
			}
			if rule.UserIDGroupPairs != nil {
				for _, sourceSecGroup := range rule.UserIDGroupPairs {
					sgRule.Spec.ForProvider.SourceSecurityGroupID = sourceSecGroup.GroupID
				}
			}

			if rule.IPRanges != nil {
				for _, ipRanges := range rule.IPRanges {
					sgRule.Spec.ForProvider.CidrBlocks = append(sgRule.Spec.ForProvider.CidrBlocks, common.PtrFromString(ipRanges.CIDRIP))
				}
			}

			matchController := true
			if len(externalNameSecurityGroupId) > 0 {
				sgRule.Spec.ForProvider.SecurityGroupID = &externalNameSecurityGroupId
			}
			sgRule.Spec.ForProvider.SecurityGroupIDRef = &runtimev1.Reference{}
			sgRule.Spec.ForProvider.SecurityGroupIDRef.Name = nameSecurityGroup
			sgRule.Spec.ForProvider.SecurityGroupIDSelector = &runtimev1.Selector{}
			sgRule.Spec.ForProvider.SecurityGroupIDSelector.MatchControllerRef = &matchController
			sgRule.Spec.ProviderConfigReference = source.Spec.ProviderConfigReference
			secGroupMRs = append(secGroupMRs, sgRule)
		}
	}
	return secGroupMRs, nil
}

func SecurityGroupComposition(sourceTemplate v1.ComposedTemplate, convertedTemplates ...*v1.ComposedTemplate) error {
	patchesToAdd, err := providerawscommon.ConvertComposedTemplateTags(sourceTemplate)
	if err != nil {
		return errors.Wrap(err, "failed to convert tags")
	}
	patchesToAdd = append(patchesToAdd, migration.ConvertComposedTemplatePatchesMap(sourceTemplate, nil)...)
	for i := range convertedTemplates {
		convertedTemplates[i].Patches = append(convertedTemplates[i].Patches, patchesToAdd...)
	}
	splittedPatchesToAdd := []v1.Patch{
		{Type: v1.PatchTypeFromCompositeFieldPath,
			FromFieldPath: common.PtrFromString("spec.forProvider.clusterSecurityGroup"),
			ToFieldPath:   common.PtrFromString("spec.forProvider.sourceSecurityGroupId"),
		},
	}
	err = common.SplittedResourcePatches(convertedTemplates, "SecurityGroupRule", splittedPatchesToAdd)
	if err != nil {
		return errors.Wrap(err, "failed to add patches to splitted resource")
	}
	return nil
}
