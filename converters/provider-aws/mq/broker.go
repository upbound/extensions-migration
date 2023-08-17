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

package mq

import (
	"fmt"
	srcv1beta1 "github.com/crossplane-contrib/provider-aws/apis/mq/v1alpha1"
	v1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/pkg/errors"
	targetv1beta1 "github.com/upbound/provider-aws/apis/mq/v1beta1"
	"github.com/upbound/upjet/pkg/migration"
	"strconv"
)

func BrokerResource(mg resource.Managed) ([]resource.Managed, error) {
	source := mg.(*srcv1beta1.Broker)
	target := &targetv1beta1.Broker{}
	brokerMRs := []resource.Managed{target}
	// TODO: spec.forProvider.ldapServerMetadata.serviceAccountPassword -> serviceAccountPasswordSecretRef
	skipFields := []string{
		"spec.forProvider.configuration",
		"spec.forProvider.creatorRequestID",
		"spec.forProvider.encryptionOptions",
		"spec.forProvider.ldapServerMetadata",
		"spec.forProvider.logs",
		"spec.forProvider.maintenanceWindowStartTime",
		"spec.forProvider.region",
		"spec.forProvider.securityGroupIdRefs",
		"spec.forProvider.securityGroupIdSelector",
		"spec.forProvider.subnetIDRefs",
		"spec.forProvider.subnetIDSelector",
		"spec.forProvider.subnetIDs",
		"spec.forProvider.users",
	}
	if _, err := migration.CopyInto(source, target, targetv1beta1.Broker_GroupVersionKind, skipFields...); err != nil {
		return nil, errors.Wrap(err, "failed to copy source into target")
	}

	if source.Spec.ForProvider.Configuration != nil {
		target.Spec.ForProvider.Configuration = make([]targetv1beta1.ConfigurationParameters, 1)
		target.Spec.ForProvider.Configuration[0].ID = source.Spec.ForProvider.Configuration.ID
		// no-op: ref types introduced at target
		// target.Spec.ForProvider.Configuration[0].IDRef
		// target.Spec.ForProvider.Configuration[0].IDSelector
		if source.Spec.ForProvider.Configuration.Revision != nil {
			// TODO: use utility function for *int64 -> *float64 conversions
			revision := float64(*source.Spec.ForProvider.Configuration.Revision)
			target.Spec.ForProvider.Configuration[0].Revision = &revision
		}
	}

	if source.Spec.ForProvider.EncryptionOptions != nil {
		target.Spec.ForProvider.EncryptionOptions = make([]targetv1beta1.EncryptionOptionsParameters, 1)
		target.Spec.ForProvider.EncryptionOptions[0].KMSKeyID = source.Spec.ForProvider.EncryptionOptions.KMSKeyID
		target.Spec.ForProvider.EncryptionOptions[0].UseAwsOwnedKey = source.Spec.ForProvider.EncryptionOptions.UseAWSOwnedKey
	}

	if source.Spec.ForProvider.LDAPServerMetadata != nil {
		target.Spec.ForProvider.LdapServerMetadata = make([]targetv1beta1.LdapServerMetadataParameters, 1)
		sourceLsm := source.Spec.ForProvider.LDAPServerMetadata
		target.Spec.ForProvider.LdapServerMetadata[0] = targetv1beta1.LdapServerMetadataParameters{
			Hosts:                           sourceLsm.Hosts,
			RoleBase:                        sourceLsm.RoleBase,
			RoleName:                        sourceLsm.RoleName,
			RoleSearchMatching:              sourceLsm.RoleSearchMatching,
			RoleSearchSubtree:               sourceLsm.RoleSearchSubtree,
			ServiceAccountPasswordSecretRef: nil,
			ServiceAccountUsername:          sourceLsm.ServiceAccountUsername,
			UserBase:                        sourceLsm.UserBase,
			UserRoleName:                    sourceLsm.UserRoleName,
			UserSearchMatching:              sourceLsm.UserSearchMatching,
			UserSearchSubtree:               sourceLsm.UserSearchSubtree,
		}
		if source.Spec.ForProvider.LDAPServerMetadata.ServiceAccountPassword != nil {
			// TODO: generate secret MR
			// TODO: deduce secret name & ns
			ldapServiceAccountSecretSelector := &v1.SecretKeySelector{
				SecretReference: v1.SecretReference{
					Name:      fmt.Sprintf("%s-ldap-sa-creds", source.Name),
					Namespace: source.Namespace,
				},
				Key: "password",
			}
			target.Spec.ForProvider.LdapServerMetadata[0].ServiceAccountPasswordSecretRef = ldapServiceAccountSecretSelector
		}
	}

	if source.Spec.ForProvider.Logs != nil {
		target.Spec.ForProvider.Logs = make([]targetv1beta1.LogsParameters, 1)
		target.Spec.ForProvider.Logs[0] = targetv1beta1.LogsParameters{}
		if source.Spec.ForProvider.Logs.Audit != nil {
			audit := strconv.FormatBool(*source.Spec.ForProvider.Logs.Audit)
			target.Spec.ForProvider.Logs[0].Audit = &audit
		}
		target.Spec.ForProvider.Logs[0].General = source.Spec.ForProvider.Logs.General
	}

	if source.Spec.ForProvider.MaintenanceWindowStartTime != nil {
		target.Spec.ForProvider.MaintenanceWindowStartTime = make([]targetv1beta1.MaintenanceWindowStartTimeParameters, 1)
		target.Spec.ForProvider.MaintenanceWindowStartTime[0] = targetv1beta1.MaintenanceWindowStartTimeParameters{
			DayOfWeek: source.Spec.ForProvider.MaintenanceWindowStartTime.DayOfWeek,
			TimeOfDay: source.Spec.ForProvider.MaintenanceWindowStartTime.TimeOfDay,
			TimeZone:  source.Spec.ForProvider.MaintenanceWindowStartTime.TimeZone,
		}
	}

	// json tag changed
	target.Spec.ForProvider.Region = &source.Spec.ForProvider.Region
	target.Spec.ForProvider.SecurityGroupRefs = source.Spec.ForProvider.SecurityGroupIDRefs
	target.Spec.ForProvider.SecurityGroupSelector = source.Spec.ForProvider.SecurityGroupIDSelector
	// json tag changed
	target.Spec.ForProvider.SubnetIDRefs = source.Spec.ForProvider.SubnetIDRefs
	target.Spec.ForProvider.SubnetIDSelector = source.Spec.ForProvider.SubnetIDSelector
	target.Spec.ForProvider.SubnetIds = source.Spec.ForProvider.SubnetIDs

	// json tag changed
	if source.Spec.ForProvider.CustomUsers != nil {
		target.Spec.ForProvider.User = make([]targetv1beta1.UserParameters, len(source.Spec.ForProvider.CustomUsers))
		for i, userParams := range source.Spec.ForProvider.CustomUsers {
			target.Spec.ForProvider.User[i] = targetv1beta1.UserParameters{
				ConsoleAccess:     userParams.ConsoleAccess,
				Groups:            userParams.Groups,
				PasswordSecretRef: userParams.PasswordSecretRef,
				Username:          userParams.Username,
			}
		}

	}

	// TODO: parameter removed at target
	// ? = source.Spec.ForProvider.CreatorRequestID

	// TODO: new parameter at target
	// target.Spec.ForProvider.ApplyImmediately
	// target.Spec.ForProvider.BrokerName

	return brokerMRs, nil
}
