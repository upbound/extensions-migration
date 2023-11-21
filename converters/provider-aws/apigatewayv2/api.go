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
	"github.com/crossplane/upjet/pkg/migration"
	"github.com/pkg/errors"
	"github.com/upbound/extensions-migration/converters/common"
	targetv1beta1 "github.com/upbound/provider-aws/apis/apigatewayv2/v1beta1"
)

func APIResource(mg resource.Managed) ([]resource.Managed, error) {
	source := mg.(*srcv1alpha1.API)
	target := &targetv1beta1.API{}
	skipFields := []string{
		"spec.forProvider.corsConfiguration",
		"spec.forProvider.credentialsARN",
		"spec.forProvider.disableExecuteAPIEndpoint",
		"spec.forProvider.disableSchemaValidation", // removed on target
	}
	if _, err := migration.CopyInto(source, target, targetv1beta1.API_GroupVersionKind, skipFields...); err != nil {
		return nil, errors.Wrap(err, "failed to copy source into target")
	}

	if source.Spec.ForProvider.CORSConfiguration != nil {
		target.Spec.ForProvider.CorsConfiguration = make([]targetv1beta1.CorsConfigurationParameters, 1)
		target.Spec.ForProvider.CorsConfiguration[0] = targetv1beta1.CorsConfigurationParameters{
			AllowCredentials: source.Spec.ForProvider.CORSConfiguration.AllowCredentials,
			AllowHeaders:     source.Spec.ForProvider.CORSConfiguration.AllowHeaders,
			AllowMethods:     source.Spec.ForProvider.CORSConfiguration.AllowMethods,
			AllowOrigins:     source.Spec.ForProvider.CORSConfiguration.AllowOrigins,
			ExposeHeaders:    source.Spec.ForProvider.CORSConfiguration.ExposeHeaders,
			MaxAge:           nil,
		}
		if source.Spec.ForProvider.CORSConfiguration.MaxAge != nil {
			target.Spec.ForProvider.CorsConfiguration[0].MaxAge = common.PtrFloat64FromInt64(source.Spec.ForProvider.CORSConfiguration.MaxAge)
		}

	}

	// json tag changed
	target.Spec.ForProvider.CredentialsArn = source.Spec.ForProvider.CredentialsARN
	target.Spec.ForProvider.DisableExecuteAPIEndpoint = source.Spec.ForProvider.DisableExecuteAPIEndpoint
	// pointer type
	target.Spec.ForProvider.Region = &source.Spec.ForProvider.Region

	// TODO: parameter removed at target
	// ? = source.Spec.ForProvider.DisableSchemaValidation
	// TODO: new parameter at target
	// target.Spec.ForProvider.Body = ?
	// target.Spec.ForProvider.FailOnWarnings = ?
	return []resource.Managed{
		target,
	}, nil
}
