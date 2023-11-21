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
	targetv1beta1 "github.com/upbound/provider-aws/apis/apigatewayv2/v1beta1"
)

func DomainNameResource(mg resource.Managed) ([]resource.Managed, error) {
	source := mg.(*srcv1alpha1.DomainName)
	target := &targetv1beta1.DomainName{}
	skipFields := []string{
		"spec.forProvider.domainNameConfigurations",
		"spec.forProvider.mutualTLSAuthentication",
		"spec.forProvider.region",
	}
	if _, err := migration.CopyInto(source, target, targetv1beta1.DomainName_GroupVersionKind, skipFields...); err != nil {
		return nil, errors.Wrap(err, "failed to copy source into target")
	}
	// object -> array
	if source.Spec.ForProvider.DomainNameConfigurations != nil {
		target.Spec.ForProvider.DomainNameConfiguration = make([]targetv1beta1.DomainNameConfigurationParameters, len(source.Spec.ForProvider.DomainNameConfigurations))
		for i, sourceDnc := range source.Spec.ForProvider.DomainNameConfigurations {
			targetDnc := targetv1beta1.DomainNameConfigurationParameters{
				CertificateArn:                      sourceDnc.CertificateARN,
				CertificateArnRef:                   nil,
				CertificateArnSelector:              nil,
				EndpointType:                        sourceDnc.EndpointType,
				OwnershipVerificationCertificateArn: sourceDnc.OwnershipVerificationCertificateARN,
				SecurityPolicy:                      sourceDnc.SecurityPolicy,
			}
			// TODO: parameter removed at target
			// sourceDnc.APIGatewayDomainName
			// sourceDnc.CertificateName
			// sourceDnc.CertificateUploadDate
			// sourceDnc.DomainNameStatus
			// sourceDnc.DomainNameStatusMessage

			// no-op: ref types introduced at target
			// targetDnc.CertificateArnRef
			// targetDnc.CertificateArnSelector

			target.Spec.ForProvider.DomainNameConfiguration[i] = targetDnc
		}
	}

	// object -> array
	if source.Spec.ForProvider.MutualTLSAuthentication != nil {
		target.Spec.ForProvider.MutualTLSAuthentication = make([]targetv1beta1.MutualTLSAuthenticationParameters, 1)
		target.Spec.ForProvider.MutualTLSAuthentication[0].TruststoreURI = source.Spec.ForProvider.MutualTLSAuthentication.TruststoreURI
		target.Spec.ForProvider.MutualTLSAuthentication[0].TruststoreVersion = source.Spec.ForProvider.MutualTLSAuthentication.TruststoreVersion
	}

	// pointer type
	target.Spec.ForProvider.Region = &source.Spec.ForProvider.Region

	return []resource.Managed{
		target,
	}, nil
}
