// Copyright 2022 Upbound Inc.
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

func DistributionResource(mg resource.Managed) ([]resource.Managed, error) {
	source := mg.(*srcv1alpha1.Distribution)
	target := &targetv1beta1.Distribution{}
	if _, err := migration.CopyInto(source, target, targetv1beta1.Distribution_GroupVersionKind, "spec.forProvider.tags"); err != nil {
		return nil, errors.Wrap(err, "failed to copy source into target")
	}

	target.Spec.ForProvider.Comment = source.Spec.ForProvider.DistributionConfig.Comment

	target.Spec.ForProvider.DefaultCacheBehavior = make([]targetv1beta1.DefaultCacheBehaviorParameters, 1)
	target.Spec.ForProvider.DefaultCacheBehavior[0].AllowedMethods = source.Spec.ForProvider.DistributionConfig.DefaultCacheBehavior.AllowedMethods.Items
	target.Spec.ForProvider.DefaultCacheBehavior[0].CachedMethods = source.Spec.ForProvider.DistributionConfig.DefaultCacheBehavior.AllowedMethods.CachedMethods.Items
	target.Spec.ForProvider.DefaultCacheBehavior[0].Compress = source.Spec.ForProvider.DistributionConfig.DefaultCacheBehavior.Compress
	target.Spec.ForProvider.DefaultCacheBehavior[0].TargetOriginID = source.Spec.ForProvider.DistributionConfig.DefaultCacheBehavior.TargetOriginID
	target.Spec.ForProvider.DefaultCacheBehavior[0].ViewerProtocolPolicy = source.Spec.ForProvider.DistributionConfig.DefaultCacheBehavior.ViewerProtocolPolicy
	target.Spec.ForProvider.DefaultCacheBehavior[0].CachePolicyID = source.Spec.ForProvider.DistributionConfig.DefaultCacheBehavior.CachePolicyID
	target.Spec.ForProvider.DefaultCacheBehavior[0].OriginRequestPolicyID = source.Spec.ForProvider.DistributionConfig.DefaultCacheBehavior.OriginRequestPolicyID
	target.Spec.ForProvider.DefaultCacheBehavior[0].ViewerProtocolPolicy = source.Spec.ForProvider.DistributionConfig.DefaultCacheBehavior.ViewerProtocolPolicy

	target.Spec.ForProvider.Enabled = source.Spec.ForProvider.DistributionConfig.Enabled

	target.Spec.ForProvider.ViewerCertificate = make([]targetv1beta1.ViewerCertificateParameters, 1)
	target.Spec.ForProvider.ViewerCertificate[0].AcmCertificateArn = source.Spec.ForProvider.DistributionConfig.ViewerCertificate.ACMCertificateARN
	target.Spec.ForProvider.ViewerCertificate[0].CloudfrontDefaultCertificate = source.Spec.ForProvider.DistributionConfig.ViewerCertificate.CloudFrontDefaultCertificate
	target.Spec.ForProvider.ViewerCertificate[0].SSLSupportMethod = source.Spec.ForProvider.DistributionConfig.ViewerCertificate.SSLSupportMethod
	target.Spec.ForProvider.ViewerCertificate[0].MinimumProtocolVersion = source.Spec.ForProvider.DistributionConfig.ViewerCertificate.MinimumProtocolVersion

	target.Spec.ForProvider.WebACLID = source.Spec.ForProvider.DistributionConfig.WebACLID

	for _, i := range source.Spec.ForProvider.DistributionConfig.Origins.Items {
		target.Spec.ForProvider.Origin = append(target.Spec.ForProvider.Origin, targetv1beta1.OriginParameters{
			DomainName:            i.DomainName,
			OriginAccessControlID: i.ID,
			S3OriginConfig:        []targetv1beta1.S3OriginConfigParameters{{OriginAccessIdentity: i.S3OriginConfig.OriginAccessIdentity}},
		})
	}

	return []resource.Managed{
		target,
	}, nil
}
