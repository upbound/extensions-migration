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

package cloudfront

import (
	srcv1alpha1 "github.com/crossplane-contrib/provider-aws/apis/cloudfront/v1alpha1"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/pkg/errors"
	targetv1beta1 "github.com/upbound/provider-aws/apis/cloudfront/v1beta1"
	"github.com/upbound/upjet/pkg/migration"
)

func ResponseHeadersPolicyResource(mg resource.Managed) ([]resource.Managed, error) {
	source := mg.(*srcv1alpha1.ResponseHeadersPolicy)
	target := &targetv1beta1.ResponseHeadersPolicy{}
	skipFields := []string{
		"spec.forProvider.responseHeadersPolicyConfig",
		"spec.forProvider.region",
	}
	if _, err := migration.CopyInto(source, target, targetv1beta1.ResponseHeadersPolicy_GroupVersionKind, skipFields...); err != nil {
		return nil, errors.Wrap(err, "failed to copy source into target")
	}

	// Source's spec.forProvider.responseHeadersPolicyConfig fields were moved to
	// Target's spec.forProvider in the new schema
	target.Spec.ForProvider.Comment = source.Spec.ForProvider.ResponseHeadersPolicyConfig.Comment
	target.Spec.ForProvider.Name = source.Spec.ForProvider.ResponseHeadersPolicyConfig.Name
	target.Spec.ForProvider.Region = &source.Spec.ForProvider.Region

	// object -> array
	if source.Spec.ForProvider.ResponseHeadersPolicyConfig.CORSConfig != nil {
		target.Spec.ForProvider.CorsConfig = make([]targetv1beta1.CorsConfigParameters, 1)
		sourceCc := source.Spec.ForProvider.ResponseHeadersPolicyConfig.CORSConfig

		targetCcp := targetv1beta1.CorsConfigParameters{
			AccessControlAllowCredentials: sourceCc.AccessControlAllowCredentials,
			AccessControlMaxAgeSec:        nil,
			OriginOverride:                sourceCc.OriginOverride,
		}

		if sourceCc.AccessControlAllowHeaders != nil {
			targetCcp.AccessControlAllowHeaders = make([]targetv1beta1.AccessControlAllowHeadersParameters, 1)
			targetCcp.AccessControlAllowHeaders[0].Items = sourceCc.AccessControlAllowHeaders.Items
		}

		if sourceCc.AccessControlAllowMethods != nil {
			targetCcp.AccessControlAllowMethods = make([]targetv1beta1.AccessControlAllowMethodsParameters, 1)
			targetCcp.AccessControlAllowMethods[0].Items = sourceCc.AccessControlAllowMethods.Items
		}

		if sourceCc.AccessControlAllowOrigins != nil {
			targetCcp.AccessControlAllowOrigins = make([]targetv1beta1.AccessControlAllowOriginsParameters, 1)
			targetCcp.AccessControlAllowOrigins[0].Items = sourceCc.AccessControlAllowOrigins.Items
		}

		if sourceCc.AccessControlExposeHeaders != nil {
			targetCcp.AccessControlExposeHeaders = make([]targetv1beta1.AccessControlExposeHeadersParameters, 1)
			targetCcp.AccessControlExposeHeaders[0].Items = sourceCc.AccessControlExposeHeaders.Items
		}
		// TODO: use utility function for *int64 -> *float64 conversions
		maxAge := float64(*sourceCc.AccessControlMaxAgeSec)
		targetCcp.AccessControlMaxAgeSec = &maxAge

		target.Spec.ForProvider.CorsConfig[0] = targetCcp
	}
	// object -> array
	if source.Spec.ForProvider.ResponseHeadersPolicyConfig.CustomHeadersConfig != nil {
		sourceChc := source.Spec.ForProvider.ResponseHeadersPolicyConfig.CustomHeadersConfig
		target.Spec.ForProvider.CustomHeadersConfig = make([]targetv1beta1.CustomHeadersConfigParameters, 1)
		target.Spec.ForProvider.CustomHeadersConfig[0].Items = make([]targetv1beta1.CustomHeadersConfigItemsParameters, len(sourceChc.Items))
		for i, chcItem := range sourceChc.Items {
			target.Spec.ForProvider.CustomHeadersConfig[0].Items[i] = targetv1beta1.CustomHeadersConfigItemsParameters{
				Header:   chcItem.Header,
				Override: chcItem.Override,
				Value:    chcItem.Value,
			}
		}
	}
	// object -> array
	if source.Spec.ForProvider.ResponseHeadersPolicyConfig.RemoveHeadersConfig != nil {
		sourceRhc := source.Spec.ForProvider.ResponseHeadersPolicyConfig.RemoveHeadersConfig
		target.Spec.ForProvider.RemoveHeadersConfig = make([]targetv1beta1.RemoveHeadersConfigParameters, 1)
		target.Spec.ForProvider.RemoveHeadersConfig[0].Items = make([]targetv1beta1.RemoveHeadersConfigItemsParameters, len(sourceRhc.Items))

		for i, rhcItem := range sourceRhc.Items {
			target.Spec.ForProvider.RemoveHeadersConfig[0].Items[i] = targetv1beta1.RemoveHeadersConfigItemsParameters{
				Header: rhcItem.Header,
			}
		}
	}
	// object -> array
	if source.Spec.ForProvider.ResponseHeadersPolicyConfig.SecurityHeadersConfig != nil {
		sourceShc := source.Spec.ForProvider.ResponseHeadersPolicyConfig.SecurityHeadersConfig
		target.Spec.ForProvider.SecurityHeadersConfig = make([]targetv1beta1.SecurityHeadersConfigParameters, 1)

		targetShcp := targetv1beta1.SecurityHeadersConfigParameters{}

		if sourceShc.ContentSecurityPolicy != nil {
			targetShcp.ContentSecurityPolicy = make([]targetv1beta1.ContentSecurityPolicyParameters, 1)
			targetShcp.ContentSecurityPolicy[0].ContentSecurityPolicy = sourceShc.ContentSecurityPolicy.ContentSecurityPolicy
			targetShcp.ContentSecurityPolicy[0].Override = sourceShc.ContentSecurityPolicy.Override
		}

		if sourceShc.ContentTypeOptions != nil {
			targetShcp.ContentTypeOptions = make([]targetv1beta1.ContentTypeOptionsParameters, 1)
			targetShcp.ContentTypeOptions[0].Override = sourceShc.ContentTypeOptions.Override
		}

		if sourceShc.FrameOptions != nil {
			targetShcp.FrameOptions = make([]targetv1beta1.FrameOptionsParameters, 1)
			targetShcp.FrameOptions[0].FrameOption = sourceShc.FrameOptions.FrameOption
			targetShcp.FrameOptions[0].Override = sourceShc.FrameOptions.Override
		}

		if sourceShc.ReferrerPolicy != nil {
			targetShcp.ReferrerPolicy = make([]targetv1beta1.ReferrerPolicyParameters, 1)
			targetShcp.ReferrerPolicy[0].ReferrerPolicy = sourceShc.ReferrerPolicy.ReferrerPolicy
			targetShcp.ReferrerPolicy[0].Override = sourceShc.ReferrerPolicy.Override
		}

		if sourceShc.StrictTransportSecurity != nil {
			targetShcp.StrictTransportSecurity = make([]targetv1beta1.StrictTransportSecurityParameters, 1)
			if sourceShc.StrictTransportSecurity.AccessControlMaxAgeSec != nil {
				// TODO: use utility function for *int64 -> *float64 conversions
				acMaxAge := float64(*sourceShc.StrictTransportSecurity.AccessControlMaxAgeSec)
				targetShcp.StrictTransportSecurity[0].AccessControlMaxAgeSec = &acMaxAge
			}
			targetShcp.StrictTransportSecurity[0].IncludeSubdomains = sourceShc.StrictTransportSecurity.IncludeSubdomains
			targetShcp.StrictTransportSecurity[0].Override = sourceShc.StrictTransportSecurity.Override
			targetShcp.StrictTransportSecurity[0].Preload = sourceShc.StrictTransportSecurity.Preload
		}

		if sourceShc.XSSProtection != nil {
			targetShcp.XSSProtection = make([]targetv1beta1.XSSProtectionParameters, 1)
			targetShcp.XSSProtection[0].ModeBlock = sourceShc.XSSProtection.ModeBlock
			targetShcp.XSSProtection[0].Override = sourceShc.XSSProtection.Override
			targetShcp.XSSProtection[0].Protection = sourceShc.XSSProtection.Protection
			targetShcp.XSSProtection[0].ReportURI = sourceShc.XSSProtection.ReportURI
		}

		target.Spec.ForProvider.SecurityHeadersConfig[0] = targetShcp

	}
	// object -> array
	if source.Spec.ForProvider.ResponseHeadersPolicyConfig.ServerTimingHeadersConfig != nil {
		sourceSthc := source.Spec.ForProvider.ResponseHeadersPolicyConfig.ServerTimingHeadersConfig
		target.Spec.ForProvider.ServerTimingHeadersConfig = make([]targetv1beta1.ServerTimingHeadersConfigParameters, 1)
		target.Spec.ForProvider.ServerTimingHeadersConfig[0].Enabled = sourceSthc.Enabled
		target.Spec.ForProvider.ServerTimingHeadersConfig[0].SamplingRate = sourceSthc.SamplingRate
	}

	// TODO: new parameter at target
	// target.Spec.ForProvider.Etag = ?

	return []resource.Managed{
		target,
	}, nil
}
