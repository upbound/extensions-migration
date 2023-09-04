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

package apigatewayv2

import (
	srcv1alpha1 "github.com/crossplane-contrib/provider-aws/apis/apigatewayv2/v1alpha1"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/pkg/errors"
	targetv1beta1 "github.com/upbound/provider-aws/apis/apigatewayv2/v1beta1"
	"github.com/upbound/upjet/pkg/migration"
)

func RouteResource(mg resource.Managed) ([]resource.Managed, error) {
	source := mg.(*srcv1alpha1.Route)
	target := &targetv1beta1.Route{}
	skipFields := []string{
		"spec.forProvider.AuthorizerID",
		"spec.forProvider.AuthorizerIDRef",
		"spec.forProvider.AuthorizerIDSelector",
		"spec.forProvider.Region",
		"spec.forProvider.RequestParameters",
	}
	if _, err := migration.CopyInto(source, target, targetv1beta1.Route_GroupVersionKind, skipFields...); err != nil {
		return nil, errors.Wrap(err, "failed to copy source into target")
	}

	// json tag changed
	target.Spec.ForProvider.AuthorizerID = source.Spec.ForProvider.AuthorizerID
	target.Spec.ForProvider.AuthorizerIDRef = source.Spec.ForProvider.AuthorizerIDRef
	target.Spec.ForProvider.AuthorizerIDSelector = source.Spec.ForProvider.AuthorizerIDSelector
	// pointer type
	target.Spec.ForProvider.Region = &source.Spec.ForProvider.Region
	// object -> array & json tag change
	if source.Spec.ForProvider.RequestParameters != nil {
		target.Spec.ForProvider.RequestParameter = make([]targetv1beta1.RequestParameterParameters, 0, len(source.Spec.ForProvider.RequestParameters))
		for k, pc := range source.Spec.ForProvider.RequestParameters {
			rp := targetv1beta1.RequestParameterParameters{
				RequestParameterKey: &k,
				Required:            pc.Required,
			}
			target.Spec.ForProvider.RequestParameter = append(target.Spec.ForProvider.RequestParameter, rp)
		}
	}

	return []resource.Managed{
		target,
	}, nil
}
