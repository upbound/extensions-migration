package provider_aws

import (
	cachev1alpha1 "github.com/crossplane-contrib/provider-aws/apis/cache/v1alpha1"
	cachev1beta1 "github.com/crossplane-contrib/provider-aws/apis/cache/v1beta1"
	cloudfrontv1alpha1 "github.com/crossplane-contrib/provider-aws/apis/cloudfront/v1alpha1"
	cloudwatchlogsv1alpha1 "github.com/crossplane-contrib/provider-aws/apis/cloudwatchlogs/v1alpha1"
	databasev1beta1 "github.com/crossplane-contrib/provider-aws/apis/database/v1beta1"
	docdbv1alpha1 "github.com/crossplane-contrib/provider-aws/apis/docdb/v1alpha1"
	dynamodbv1alpha1 "github.com/crossplane-contrib/provider-aws/apis/dynamodb/v1alpha1"
	ec2v1alpha1 "github.com/crossplane-contrib/provider-aws/apis/ec2/v1alpha1"
	ec2v1beta1 "github.com/crossplane-contrib/provider-aws/apis/ec2/v1beta1"
	efsv1alpha1 "github.com/crossplane-contrib/provider-aws/apis/efs/v1alpha1"
	eksmanualv1alpha1 "github.com/crossplane-contrib/provider-aws/apis/eks/manualv1alpha1"
	eksv1alpha1 "github.com/crossplane-contrib/provider-aws/apis/eks/v1alpha1"
	eksv1beta1 "github.com/crossplane-contrib/provider-aws/apis/eks/v1beta1"
	elasticachev1alpha1 "github.com/crossplane-contrib/provider-aws/apis/elasticache/v1alpha1"
	iamv1beta1 "github.com/crossplane-contrib/provider-aws/apis/iam/v1beta1"
	kmsv1alpha1 "github.com/crossplane-contrib/provider-aws/apis/kms/v1alpha1"
	rdsv1alpha1 "github.com/crossplane-contrib/provider-aws/apis/rds/v1alpha1"
	route53v1alpha1 "github.com/crossplane-contrib/provider-aws/apis/route53/v1alpha1"
	route53resolvermanualv1alpha1 "github.com/crossplane-contrib/provider-aws/apis/route53resolver/manualv1alpha1"
	s3v1alpha3 "github.com/crossplane-contrib/provider-aws/apis/s3/v1alpha3"
	s3v1beta1 "github.com/crossplane-contrib/provider-aws/apis/s3/v1beta1"
	secretsmanagerv1beta1 "github.com/crossplane-contrib/provider-aws/apis/secretsmanager/v1beta1"
	snsv1beta1 "github.com/crossplane-contrib/provider-aws/apis/sns/v1beta1"
	sqsv1beta1 "github.com/crossplane-contrib/provider-aws/apis/sqs/v1beta1"
	"github.com/upbound/extensions-migration/converters/provider-aws/cloudfront"
	"github.com/upbound/extensions-migration/converters/provider-aws/cloudwatchlogs"
	"github.com/upbound/extensions-migration/converters/provider-aws/docdb"
	"github.com/upbound/extensions-migration/converters/provider-aws/dynamodb"
	"github.com/upbound/extensions-migration/converters/provider-aws/ec2"
	"github.com/upbound/extensions-migration/converters/provider-aws/efs"
	"github.com/upbound/extensions-migration/converters/provider-aws/eks"
	"github.com/upbound/extensions-migration/converters/provider-aws/elasticache"
	"github.com/upbound/extensions-migration/converters/provider-aws/iam"
	"github.com/upbound/extensions-migration/converters/provider-aws/kms"
	"github.com/upbound/extensions-migration/converters/provider-aws/rds"
	"github.com/upbound/extensions-migration/converters/provider-aws/route53"
	"github.com/upbound/extensions-migration/converters/provider-aws/route53resolver"
	"github.com/upbound/extensions-migration/converters/provider-aws/s3"
	"github.com/upbound/extensions-migration/converters/provider-aws/secretsmanager"
	"github.com/upbound/extensions-migration/converters/provider-aws/sns"
	"github.com/upbound/extensions-migration/converters/provider-aws/sqs"
	"github.com/upbound/upjet/pkg/migration"
)

func RegisterAllKnownConverters(r *migration.Registry) {
	r.RegisterAPIConversionFunctions(cloudfrontv1alpha1.DistributionGroupVersionKind,
		cloudfront.DistributionResource, nil, migration.DefaultPatchSetsConverter)
	r.RegisterAPIConversionFunctions(cloudwatchlogsv1alpha1.LogGroupGroupVersionKind,
		cloudwatchlogs.LogGroupResource, nil, nil)
	r.RegisterAPIConversionFunctions(docdbv1alpha1.DBClusterGroupVersionKind,
		docdb.ClusterResource, migration.DefaultCompositionConverter(true, nil), migration.DefaultPatchSetsConverter)
	r.RegisterAPIConversionFunctions(docdbv1alpha1.DBClusterParameterGroupGroupVersionKind,
		docdb.ClusterParameterGroupResource, migration.DefaultCompositionConverter(true, nil), migration.DefaultPatchSetsConverter)
	r.RegisterAPIConversionFunctions(docdbv1alpha1.DBInstanceGroupVersionKind,
		docdb.InstanceResource, migration.DefaultCompositionConverter(true, nil), migration.DefaultPatchSetsConverter)
	r.RegisterAPIConversionFunctions(docdbv1alpha1.DBSubnetGroupGroupVersionKind,
		docdb.SubnetGroupResource, migration.DefaultCompositionConverter(true, nil), migration.DefaultPatchSetsConverter)
	r.RegisterAPIConversionFunctions(dynamodbv1alpha1.TableGroupVersionKind,
		dynamodb.TableResource, migration.DefaultCompositionConverter(true, nil), migration.DefaultPatchSetsConverter)
	r.RegisterAPIConversionFunctions(ec2v1alpha1.FlowLogGroupVersionKind,
		ec2.FlowLogResource, migration.DefaultCompositionConverter(true, nil), migration.DefaultPatchSetsConverter)
	r.RegisterAPIConversionFunctions(ec2v1beta1.NATGatewayGroupVersionKind,
		ec2.NATGatewayResource, migration.DefaultCompositionConverter(true, nil), migration.DefaultPatchSetsConverter)
	r.RegisterAPIConversionFunctions(ec2v1beta1.RouteTableGroupVersionKind,
		ec2.RouteTableResource, nil, nil)
	r.RegisterAPIConversionFunctions(ec2v1beta1.SecurityGroupGroupVersionKind,
		ec2.SecurityGroupResource, ec2.SecurityGroupComposition, migration.DefaultPatchSetsConverter)
	r.RegisterAPIConversionFunctions(ec2v1beta1.SubnetGroupVersionKind,
		ec2.SubnetResource, migration.DefaultCompositionConverter(true, nil), migration.DefaultPatchSetsConverter)
	r.RegisterAPIConversionFunctions(ec2v1alpha1.TransitGatewayVPCAttachmentGroupVersionKind,
		ec2.TransitGatewayVPCAttachmentResource, migration.DefaultCompositionConverter(true, nil), migration.DefaultPatchSetsConverter)
	r.RegisterAPIConversionFunctions(ec2v1beta1.VPCGroupVersionKind,
		ec2.VPCResource, migration.DefaultCompositionConverter(true, nil), migration.DefaultPatchSetsConverter)
	r.RegisterAPIConversionFunctions(ec2v1beta1.VPCCIDRBlockGroupVersionKind,
		ec2.VPCCidrBlockResource, nil, nil)
	r.RegisterAPIConversionFunctions(ec2v1alpha1.VPCEndpointGroupVersionKind,
		ec2.VPCEndpointResource, nil, nil)
	r.RegisterAPIConversionFunctions(efsv1alpha1.FileSystemGroupVersionKind,
		efs.FileSystemResource, migration.DefaultCompositionConverter(true, map[string]string{
			"spec.forProvider.kmsKeyID":       "spec.forProvider.kmsKeyId",
			"status.atProvider.fileSystemID":  "status.atProvider.id",
			"status.atProvider.fileSystemARN": "status.atProvider.arn",
		}), migration.DefaultPatchSetsConverter)
	r.RegisterAPIConversionFunctions(efsv1alpha1.MountTargetGroupVersionKind,
		efs.MountTargetResource, migration.DefaultCompositionConverter(true, map[string]string{
			"spec.forProvider.fileSystemID":         "spec.forProvider.fileSystemId",
			"spec.forProvider.fileSystemIDRef":      "spec.forProvider.fileSystemIdRef",
			"spec.forProvider.fileSystemIDSelector": "spec.forProvider.fileSystemIdSelector",
			"spec.forProvider.subnetID":             "spec.forProvider.subnetId",
			"spec.forProvider.subnetIDRef":          "spec.forProvider.subnetIdRef",
			"spec.forProvider.subnetIDSelector":     "spec.forProvider.subnetIdSelector",
		}), migration.DefaultPatchSetsConverter)
	r.RegisterAPIConversionFunctions(eksv1alpha1.AddonGroupVersionKind,
		eks.AddonResource, nil, nil)
	r.RegisterAPIConversionFunctions(eksv1beta1.ClusterGroupVersionKind,
		eks.ClusterResource, nil, nil)
	r.RegisterAPIConversionFunctions(eksmanualv1alpha1.NodeGroupGroupVersionKind,
		eks.NodegroupResource, nil, nil)
	r.RegisterAPIConversionFunctions(cachev1alpha1.CacheClusterGroupVersionKind,
		elasticache.ClusterResource, migration.DefaultCompositionConverter(true, map[string]string{
			"spec.forProvider.cacheParameterGroupName": "spec.forProvider.parameterGroupName",
		}), migration.DefaultPatchSetsConverter)
	r.RegisterAPIConversionFunctions(elasticachev1alpha1.CacheParameterGroupGroupVersionKind,
		elasticache.ParameterGroupResource, migration.DefaultCompositionConverter(true, map[string]string{
			"spec.forProvider.cacheParameterGroupFamily": "spec.forProvider.family",
		}), migration.DefaultPatchSetsConverter)
	r.RegisterAPIConversionFunctions(cachev1beta1.ReplicationGroupGroupVersionKind,
		elasticache.ReplicationGroupResource, migration.DefaultCompositionConverter(true, map[string]string{
			"spec.forProvider.cacheParameterGroupName": "spec.forProvider.parameterGroupName",
		}), migration.DefaultPatchSetsConverter)
	r.RegisterAPIConversionFunctions(iamv1beta1.PolicyGroupVersionKind,
		iam.PolicyResource, migration.DefaultCompositionConverter(true, map[string]string{
			"spec.forProvider.document": "spec.forProvider.policy",
		}), migration.DefaultPatchSetsConverter)
	r.RegisterAPIConversionFunctions(iamv1beta1.RoleGroupVersionKind,
		iam.RoleResource, migration.DefaultCompositionConverter(true, map[string]string{
			"spec.forProvider.assumeRolePolicyDocument": "spec.forProvider.assumeRolePolicy",
		}), migration.DefaultPatchSetsConverter)
	r.RegisterAPIConversionFunctions(iamv1beta1.RolePolicyAttachmentGroupVersionKind,
		iam.RolePolicyAttachmentResource, migration.DefaultCompositionConverter(false, map[string]string{
			"spec.forProvider.roleName":         "spec.forProvider.role",
			"spec.forProvider.roleNameRef":      "spec.forProvider.roleRef",
			"spec.forProvider.roleNameSelector": "spec.forProvider.roleSelector",
		}), migration.DefaultPatchSetsConverter)
	r.RegisterAPIConversionFunctions(kmsv1alpha1.AliasGroupVersionKind,
		kms.AliasResource, nil, nil)
	r.RegisterAPIConversionFunctions(kmsv1alpha1.KeyGroupVersionKind,
		kms.KeyResource, migration.DefaultCompositionConverter(true, map[string]string{
			"spec.forProvider.enabled": "spec.forProvider.isEnabled",
			"status.atProvider.keyID":  "status.atProvider.keyId",
		}), migration.DefaultPatchSetsConverter)
	r.RegisterAPIConversionFunctions(rdsv1alpha1.DBParameterGroupGroupVersionKind,
		rds.ParameterGroupResource, migration.DefaultCompositionConverter(true, nil), migration.DefaultPatchSetsConverter)
	r.RegisterAPIConversionFunctions(databasev1beta1.RDSInstanceGroupVersionKind,
		rds.InstanceResource, migration.DefaultCompositionConverter(true, nil), migration.DefaultPatchSetsConverter)
	r.RegisterAPIConversionFunctions(databasev1beta1.DBSubnetGroupGroupVersionKind,
		rds.DBSubnetGroupResource, migration.DefaultCompositionConverter(true, nil), migration.DefaultPatchSetsConverter)
	r.RegisterAPIConversionFunctions(route53v1alpha1.HostedZoneGroupVersionKind,
		route53.HostedZoneResource, migration.DefaultCompositionConverter(true, nil), migration.DefaultPatchSetsConverter)
	r.RegisterAPIConversionFunctions(route53v1alpha1.ResourceRecordSetGroupVersionKind,
		route53.ResourceRecordSetResource, nil, nil)
	r.RegisterAPIConversionFunctions(route53resolvermanualv1alpha1.ResolverRuleAssociationGroupVersionKind,
		route53resolver.ResolverRuleAssociationResource, nil, nil)
	r.RegisterAPIConversionFunctions(s3v1alpha3.BucketPolicyGroupVersionKind,
		s3.BucketPolicyResource, nil, nil)
	r.RegisterAPIConversionFunctions(s3v1beta1.BucketGroupVersionKind,
		s3.BucketResource, nil, nil)
	r.RegisterAPIConversionFunctions(secretsmanagerv1beta1.SecretGroupVersionKind,
		secretsmanager.SecretResource, migration.DefaultCompositionConverter(true, map[string]string{
			"spec.forProvider.assumeRolePolicyDocument": "spec.forProvider.assumeRolePolicy",
		}), migration.DefaultPatchSetsConverter)
	r.RegisterAPIConversionFunctions(snsv1beta1.SubscriptionGroupVersionKind,
		sns.SubscriptionResource, migration.DefaultCompositionConverter(true, nil), migration.DefaultPatchSetsConverter)
	r.RegisterAPIConversionFunctions(snsv1beta1.TopicGroupVersionKind,
		sns.TopicResource, migration.DefaultCompositionConverter(true, nil), migration.DefaultPatchSetsConverter)
	r.RegisterAPIConversionFunctions(sqsv1beta1.QueueGroupVersionKind,
		sqs.QueueResource, migration.DefaultCompositionConverter(true, nil), migration.DefaultPatchSetsConverter)
}
