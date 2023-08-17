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

func StageResource(mg resource.Managed) ([]resource.Managed, error) {
	source := mg.(*srcv1alpha1.Stage)
	target := &targetv1beta1.Stage{}
	skipFields := []string{
		"spec.forProvider.accessLogSettings",
		"spec.forProvider.clientCertificateId",
		"spec.forProvider.defaultRouteSettings",
		"spec.forProvider.deploymentID",
	}
	if _, err := migration.CopyInto(source, target, targetv1beta1.Stage_GroupVersionKind, skipFields...); err != nil {
		return nil, errors.Wrap(err, "failed to copy source into target")
	}

	// object -> array
	if source.Spec.ForProvider.AccessLogSettings != nil {
		target.Spec.ForProvider.AccessLogSettings = make([]targetv1beta1.AccessLogSettingsParameters, 1)
		target.Spec.ForProvider.AccessLogSettings[0].DestinationArn = source.Spec.ForProvider.AccessLogSettings.DestinationARN
		target.Spec.ForProvider.AccessLogSettings[0].Format = source.Spec.ForProvider.AccessLogSettings.Format
	}
	// json tag changed
	target.Spec.ForProvider.ClientCertificateID = source.Spec.ForProvider.ClientCertificateID
	// object -> array
	if source.Spec.ForProvider.DefaultRouteSettings != nil {
		target.Spec.ForProvider.DefaultRouteSettings = make([]targetv1beta1.DefaultRouteSettingsParameters, 1)
		target.Spec.ForProvider.DefaultRouteSettings[0].DataTraceEnabled = source.Spec.ForProvider.DefaultRouteSettings.DataTraceEnabled
		target.Spec.ForProvider.DefaultRouteSettings[0].DetailedMetricsEnabled = source.Spec.ForProvider.DefaultRouteSettings.DetailedMetricsEnabled
		target.Spec.ForProvider.DefaultRouteSettings[0].LoggingLevel = source.Spec.ForProvider.DefaultRouteSettings.LoggingLevel
		// TODO: use utility function for *int64 -> *float64 conversions
		throttlingBurstLimit := float64(*source.Spec.ForProvider.DefaultRouteSettings.ThrottlingBurstLimit)
		target.Spec.ForProvider.DefaultRouteSettings[0].ThrottlingBurstLimit = &throttlingBurstLimit
		target.Spec.ForProvider.DefaultRouteSettings[0].ThrottlingRateLimit = source.Spec.ForProvider.DefaultRouteSettings.ThrottlingRateLimit
	}

	// json tag changed
	target.Spec.ForProvider.DeploymentID = source.Spec.ForProvider.DeploymentID

	// no-op: ref types introduced at target
	// target.Spec.ForProvider.DeploymentIDRef = source.Spec.ForProvider.DeploymentIDRef
	// target.Spec.ForProvider.DeploymentIDSelector = source.Spec.ForProvider.DeploymentIDSelector

	// pointer type
	target.Spec.ForProvider.Region = &source.Spec.ForProvider.Region

	// object -> array
	if source.Spec.ForProvider.RouteSettings != nil {
		target.Spec.ForProvider.RouteSettings = make([]targetv1beta1.RouteSettingsParameters, 0, len(source.Spec.ForProvider.RouteSettings))
		for routeKey, rs := range source.Spec.ForProvider.RouteSettings {
			targetRsp := targetv1beta1.RouteSettingsParameters{
				DataTraceEnabled:       rs.DataTraceEnabled,
				DetailedMetricsEnabled: rs.DetailedMetricsEnabled,
				LoggingLevel:           rs.LoggingLevel,
				RouteKey:               &routeKey,
				ThrottlingBurstLimit:   nil,
				ThrottlingRateLimit:    rs.ThrottlingRateLimit,
			}
			// TODO: use utility function for *int64 -> *float64 conversions
			burstLimit := float64(*rs.ThrottlingBurstLimit)
			targetRsp.ThrottlingBurstLimit = &burstLimit
			target.Spec.ForProvider.RouteSettings = append(target.Spec.ForProvider.RouteSettings, targetRsp)
		}
	}

	return []resource.Managed{
		target,
	}, nil
}
