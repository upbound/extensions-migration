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

package kafka

import (
	srcv1alpha1 "github.com/crossplane-contrib/provider-aws/apis/kafka/v1alpha1"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/crossplane/upjet/pkg/migration"
	"github.com/pkg/errors"
	"github.com/upbound/extensions-migration/converters/common"
	targetv1beta1 "github.com/upbound/provider-aws/apis/kafka/v1beta1"
)

func ClusterResource(mg resource.Managed) ([]resource.Managed, error) {
	source := mg.(*srcv1alpha1.Cluster)
	target := &targetv1beta1.Cluster{}
	if _, err := migration.CopyInto(source, target, targetv1beta1.Cluster_GroupVersionKind,
		"spec.forProvider.configurationInfo", "spec.forProvider.brokerNodeGroupInfo", "spec.forProvider.clientAuthentication",
		"spec.forProvider.encryptionInfo", "spec.forProvider.loggingInfo", "spec.forProvider.openMonitoring"); err != nil {
		return nil, errors.Wrap(err, "failed to copy source into target")
	}
	target.Spec.ForProvider.Region = &source.Spec.ForProvider.Region
	if source.Spec.ForProvider.CustomConfigurationInfo != nil {
		target.Spec.ForProvider.ConfigurationInfo = make([]targetv1beta1.ConfigurationInfoParameters, 1)
		target.Spec.ForProvider.ConfigurationInfo[0].Arn = source.Spec.ForProvider.CustomConfigurationInfo.ARN
		target.Spec.ForProvider.ConfigurationInfo[0].Revision = common.PtrFloat64FromInt64(source.Spec.ForProvider.CustomConfigurationInfo.Revision)
	}
	if source.Spec.ForProvider.CustomBrokerNodeGroupInfo != nil {
		target.Spec.ForProvider.BrokerNodeGroupInfo = make([]targetv1beta1.BrokerNodeGroupInfoParameters, 1)
		target.Spec.ForProvider.BrokerNodeGroupInfo[0].ClientSubnets = source.Spec.ForProvider.CustomBrokerNodeGroupInfo.ClientSubnets
		target.Spec.ForProvider.BrokerNodeGroupInfo[0].ClientSubnetsRefs = source.Spec.ForProvider.CustomBrokerNodeGroupInfo.ClientSubnetRefs
		target.Spec.ForProvider.BrokerNodeGroupInfo[0].ClientSubnetsSelector = source.Spec.ForProvider.CustomBrokerNodeGroupInfo.ClientSubnetSelector
		target.Spec.ForProvider.BrokerNodeGroupInfo[0].InstanceType = source.Spec.ForProvider.CustomBrokerNodeGroupInfo.InstanceType
		target.Spec.ForProvider.BrokerNodeGroupInfo[0].SecurityGroups = source.Spec.ForProvider.CustomBrokerNodeGroupInfo.SecurityGroups
		target.Spec.ForProvider.BrokerNodeGroupInfo[0].SecurityGroupsRefs = source.Spec.ForProvider.CustomBrokerNodeGroupInfo.SecurityGroupRefs
		target.Spec.ForProvider.BrokerNodeGroupInfo[0].SecurityGroupsSelector = source.Spec.ForProvider.CustomBrokerNodeGroupInfo.SecurityGroupSelector
		if source.Spec.ForProvider.CustomBrokerNodeGroupInfo.StorageInfo != nil {
			target.Spec.ForProvider.BrokerNodeGroupInfo[0].StorageInfo = make([]targetv1beta1.StorageInfoParameters, 1)
			if source.Spec.ForProvider.CustomBrokerNodeGroupInfo.StorageInfo.EBSStorageInfo != nil {
				target.Spec.ForProvider.BrokerNodeGroupInfo[0].StorageInfo[0].EBSStorageInfo = make([]targetv1beta1.EBSStorageInfoParameters, 1)
				target.Spec.ForProvider.BrokerNodeGroupInfo[0].StorageInfo[0].EBSStorageInfo[0].VolumeSize = common.PtrFloat64FromInt64(source.Spec.ForProvider.CustomBrokerNodeGroupInfo.StorageInfo.EBSStorageInfo.VolumeSize)
				if source.Spec.ForProvider.CustomBrokerNodeGroupInfo.StorageInfo.EBSStorageInfo.ProvisionedThroughput != nil {
					target.Spec.ForProvider.BrokerNodeGroupInfo[0].StorageInfo[0].EBSStorageInfo[0].ProvisionedThroughput = make([]targetv1beta1.ProvisionedThroughputParameters, 1)
					target.Spec.ForProvider.BrokerNodeGroupInfo[0].StorageInfo[0].EBSStorageInfo[0].ProvisionedThroughput[0].VolumeThroughput = common.PtrFloat64FromInt64(source.Spec.ForProvider.CustomBrokerNodeGroupInfo.StorageInfo.EBSStorageInfo.ProvisionedThroughput.VolumeThroughput)
					target.Spec.ForProvider.BrokerNodeGroupInfo[0].StorageInfo[0].EBSStorageInfo[0].ProvisionedThroughput[0].Enabled = source.Spec.ForProvider.CustomBrokerNodeGroupInfo.StorageInfo.EBSStorageInfo.ProvisionedThroughput.Enabled
				}
			}
		}

		// NOTE: The following fields are newly introduced at the target API
		// target.Spec.ForProvider.BrokerNodeGroupInfo[0].AzDistribution
		// target.Spec.ForProvider.BrokerNodeGroupInfo[0].EBSVolumeSize
	}
	if source.Spec.ForProvider.ClientAuthentication != nil {
		target.Spec.ForProvider.ClientAuthentication = make([]targetv1beta1.ClientAuthenticationParameters, 1)
		target.Spec.ForProvider.ClientAuthentication[0].Unauthenticated = source.Spec.ForProvider.ClientAuthentication.Unauthenticated.Enabled
		if source.Spec.ForProvider.ClientAuthentication.TLS != nil {
			target.Spec.ForProvider.ClientAuthentication[0].TLS = make([]targetv1beta1.TLSParameters, 1)
			target.Spec.ForProvider.ClientAuthentication[0].TLS[0].CertificateAuthorityArns = source.Spec.ForProvider.ClientAuthentication.TLS.CertificateAuthorityARNList
		}
		if source.Spec.ForProvider.ClientAuthentication.SASL != nil {
			target.Spec.ForProvider.ClientAuthentication[0].Sasl = make([]targetv1beta1.SaslParameters, 1)
			target.Spec.ForProvider.ClientAuthentication[0].Sasl[0].IAM = source.Spec.ForProvider.ClientAuthentication.SASL.IAM.Enabled
			target.Spec.ForProvider.ClientAuthentication[0].Sasl[0].Scram = source.Spec.ForProvider.ClientAuthentication.SASL.SCRAM.Enabled
		}
	}
	if source.Spec.ForProvider.EncryptionInfo != nil {
		target.Spec.ForProvider.EncryptionInfo = make([]targetv1beta1.EncryptionInfoParameters, 1)
		target.Spec.ForProvider.EncryptionInfo[0].EncryptionAtRestKMSKeyArn = source.Spec.ForProvider.EncryptionInfo.EncryptionAtRest.DataVolumeKMSKeyID
		if source.Spec.ForProvider.EncryptionInfo.EncryptionInTransit != nil {
			target.Spec.ForProvider.EncryptionInfo[0].EncryptionInTransit = make([]targetv1beta1.EncryptionInTransitParameters, 1)
			target.Spec.ForProvider.EncryptionInfo[0].EncryptionInTransit[0].InCluster = source.Spec.ForProvider.EncryptionInfo.EncryptionInTransit.InCluster
			target.Spec.ForProvider.EncryptionInfo[0].EncryptionInTransit[0].ClientBroker = source.Spec.ForProvider.EncryptionInfo.EncryptionInTransit.ClientBroker
		}

		// NOTE: EncryptionAtRestKMSKeyArn started to accept xp reference and selector in the new API
		// target.Spec.ForProvider.EncryptionInfo[0].EncryptionAtRestKMSKeyArnRef
		// target.Spec.ForProvider.EncryptionInfo[0].EncryptionAtRestKMSKeyArnSelector
	}
	if source.Spec.ForProvider.LoggingInfo != nil {
		target.Spec.ForProvider.LoggingInfo = make([]targetv1beta1.LoggingInfoParameters, 1)
		if source.Spec.ForProvider.LoggingInfo.BrokerLogs != nil {
			target.Spec.ForProvider.LoggingInfo[0].BrokerLogs = make([]targetv1beta1.BrokerLogsParameters, 1)
			if source.Spec.ForProvider.LoggingInfo.BrokerLogs.CloudWatchLogs != nil {
				target.Spec.ForProvider.LoggingInfo[0].BrokerLogs[0].CloudwatchLogs = make([]targetv1beta1.CloudwatchLogsParameters, 1)
				target.Spec.ForProvider.LoggingInfo[0].BrokerLogs[0].CloudwatchLogs[0].LogGroup = source.Spec.ForProvider.LoggingInfo.BrokerLogs.CloudWatchLogs.LogGroup
				target.Spec.ForProvider.LoggingInfo[0].BrokerLogs[0].CloudwatchLogs[0].Enabled = source.Spec.ForProvider.LoggingInfo.BrokerLogs.CloudWatchLogs.Enabled
			}
			if source.Spec.ForProvider.LoggingInfo.BrokerLogs.S3 != nil {
				target.Spec.ForProvider.LoggingInfo[0].BrokerLogs[0].S3 = make([]targetv1beta1.S3Parameters, 1)
				target.Spec.ForProvider.LoggingInfo[0].BrokerLogs[0].S3[0].Bucket = source.Spec.ForProvider.LoggingInfo.BrokerLogs.S3.Bucket
				target.Spec.ForProvider.LoggingInfo[0].BrokerLogs[0].S3[0].Enabled = source.Spec.ForProvider.LoggingInfo.BrokerLogs.S3.Enabled
				target.Spec.ForProvider.LoggingInfo[0].BrokerLogs[0].S3[0].Prefix = source.Spec.ForProvider.LoggingInfo.BrokerLogs.S3.Prefix
			}
			if source.Spec.ForProvider.LoggingInfo.BrokerLogs.Firehose != nil {
				target.Spec.ForProvider.LoggingInfo[0].BrokerLogs[0].Firehose = make([]targetv1beta1.FirehoseParameters, 1)
				target.Spec.ForProvider.LoggingInfo[0].BrokerLogs[0].Firehose[0].DeliveryStream = source.Spec.ForProvider.LoggingInfo.BrokerLogs.Firehose.DeliveryStream
				target.Spec.ForProvider.LoggingInfo[0].BrokerLogs[0].Firehose[0].Enabled = source.Spec.ForProvider.LoggingInfo.BrokerLogs.Firehose.Enabled
			}
		}
	}
	if source.Spec.ForProvider.OpenMonitoring != nil {
		target.Spec.ForProvider.OpenMonitoring = make([]targetv1beta1.OpenMonitoringParameters, 1)
		if source.Spec.ForProvider.OpenMonitoring.Prometheus != nil {
			target.Spec.ForProvider.OpenMonitoring[0].Prometheus = make([]targetv1beta1.PrometheusParameters, 1)
			if source.Spec.ForProvider.OpenMonitoring.Prometheus.JmxExporter != nil {
				target.Spec.ForProvider.OpenMonitoring[0].Prometheus[0].JmxExporter = make([]targetv1beta1.JmxExporterParameters, 1)
				target.Spec.ForProvider.OpenMonitoring[0].Prometheus[0].JmxExporter[0].EnabledInBroker = source.Spec.ForProvider.OpenMonitoring.Prometheus.JmxExporter.EnabledInBroker
			}
			if source.Spec.ForProvider.OpenMonitoring.Prometheus.NodeExporter != nil {
				target.Spec.ForProvider.OpenMonitoring[0].Prometheus[0].NodeExporter = make([]targetv1beta1.NodeExporterParameters, 1)
				target.Spec.ForProvider.OpenMonitoring[0].Prometheus[0].NodeExporter[0].EnabledInBroker = source.Spec.ForProvider.OpenMonitoring.Prometheus.NodeExporter.EnabledInBroker
			}
		}
	}

	return []resource.Managed{
		target,
	}, nil
}
