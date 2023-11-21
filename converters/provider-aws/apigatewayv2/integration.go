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
	"fmt"

	srcv1alpha1 "github.com/crossplane-contrib/provider-aws/apis/apigatewayv2/v1alpha1"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/crossplane/upjet/pkg/migration"
	"github.com/pkg/errors"
	"github.com/upbound/extensions-migration/converters/common"
	targetv1beta1 "github.com/upbound/provider-aws/apis/apigatewayv2/v1beta1"
)

func IntegrationResource(mg resource.Managed) ([]resource.Managed, error) {
	source := mg.(*srcv1alpha1.Integration)
	target := &targetv1beta1.Integration{}
	skipFields := []string{
		"spec.forProvider.connectionID",
		"spec.forProvider.credentialsARN",
		"spec.forProvider.integrationURI",
		"spec.forProvider.region",
		"spec.forProvider.responseParameters",
		"spec.forProvider.timeoutInMillis",
		"spec.forProvider.tlsConfig",
	}
	if _, err := migration.CopyInto(source, target, targetv1beta1.Integration_GroupVersionKind, skipFields...); err != nil {
		return nil, errors.Wrap(err, "failed to copy source into target")
	}
	// json tag changed
	target.Spec.ForProvider.ConnectionID = source.Spec.ForProvider.ConnectionID
	target.Spec.ForProvider.CredentialsArn = source.Spec.ForProvider.CredentialsARN
	// no-op: ref types introduced at target
	// target.Spec.ForProvider.CredentialsArnSelector = ?
	// target.Spec.ForProvider.CredentialsArnRef = ?

	// json tag changed
	target.Spec.ForProvider.IntegrationURI = source.Spec.ForProvider.IntegrationURI
	// no-op: ref types introduced at target
	// target.Spec.ForProvider.IntegrationURISelector = ?
	// target.Spec.ForProvider.IntegrationURIRef = ?

	target.Spec.ForProvider.Region = &source.Spec.ForProvider.Region

	// object -> array & api changes
	if source.Spec.ForProvider.ResponseParameters != nil {
		target.Spec.ForProvider.ResponseParameters = make([]targetv1beta1.ResponseParametersParameters, 0, len(source.Spec.ForProvider.ResponseParameters))
		// in community aws-provider, statuscode -> response parameter mappings are stored in a map where
		// status code is the map key and parameters struct is the value
		// the parameter struct has two fields for configuring header entries and the status code overwrite
		//
		// in the new official providers, statuscode -> response parameter mappings are stored in a struct consisting of
		// StatusCode and Mappings fields.
		// Mappings are now represented in a go map, with a special syntax supporting more transform operations compared
		// to community providers
		// - transform operation string (with the special syntax) being the map key
		// - desired value being the map value
		// see https://docs.aws.amazon.com/apigateway/latest/developerguide/http-api-parameter-mapping.html
		// for supported transform operation keys and how to construct them
		for statusCode, sourceRespParam := range source.Spec.ForProvider.ResponseParameters {
			// preserve nil in the target when HeaderEntries or OverwriteStatusCode is nil
			// this differentiates nil and empty mapping
			var mappings map[string]*string
			if sourceRespParam.HeaderEntries != nil || sourceRespParam.OverwriteStatusCode != nil {
				mappings = map[string]*string{}
				if sourceRespParam.HeaderEntries != nil {
					for _, entry := range sourceRespParam.HeaderEntries {
						operationWithName := fmt.Sprintf("%s:header.%s", entry.Operation, entry.Name)
						mappings[operationWithName] = &entry.Value
					}
				}
				if sourceRespParam.OverwriteStatusCode != nil {
					mappings["overwrite:statuscode"] = sourceRespParam.OverwriteStatusCode
				}
			}
			rpp := targetv1beta1.ResponseParametersParameters{
				Mappings:   mappings,
				StatusCode: &statusCode,
			}
			target.Spec.ForProvider.ResponseParameters = append(target.Spec.ForProvider.ResponseParameters, rpp)
		}
	}

	// type conversion
	if source.Spec.ForProvider.TimeoutInMillis != nil {
		target.Spec.ForProvider.TimeoutMilliseconds = common.PtrFloat64FromInt64(source.Spec.ForProvider.TimeoutInMillis)
	}

	// object -> array
	if source.Spec.ForProvider.TLSConfig != nil {
		target.Spec.ForProvider.TLSConfig = make([]targetv1beta1.TLSConfigParameters, 1)
		target.Spec.ForProvider.TLSConfig[0] = targetv1beta1.TLSConfigParameters{
			ServerNameToVerify: source.Spec.ForProvider.TLSConfig.ServerNameToVerify,
		}
	}

	return []resource.Managed{
		target,
	}, nil
}
