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

package route53resolver

import (
	srcv1alpha1 "github.com/crossplane-contrib/provider-aws/apis/route53resolver/manualv1alpha1"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/crossplane/upjet/pkg/migration"
	"github.com/pkg/errors"
	targetv1beta1 "github.com/upbound/provider-aws/apis/route53resolver/v1beta1"
)

func ResolverRuleAssociationResource(mg resource.Managed) ([]resource.Managed, error) {
	source := mg.(*srcv1alpha1.ResolverRuleAssociation)
	target := &targetv1beta1.RuleAssociation{}
	if _, err := migration.CopyInto(source, target, targetv1beta1.RuleAssociation_GroupVersionKind); err != nil {
		return nil, errors.Wrap(err, "failed to copy source into target")
	}
	return []resource.Managed{
		target,
	}, nil
}
