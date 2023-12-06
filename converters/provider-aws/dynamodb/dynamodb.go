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

package dynamodb

import (
	srcv1alpha1 "github.com/crossplane-contrib/provider-aws/apis/dynamodb/v1alpha1"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/crossplane/upjet/pkg/migration"
	"github.com/pkg/errors"
	targetv1beta1 "github.com/upbound/provider-aws/apis/dynamodb/v1beta1"
)

func TableResource(mg resource.Managed) ([]resource.Managed, error) {
	source := mg.(*srcv1alpha1.Table)
	target := &targetv1beta1.Table{}
	if _, err := migration.CopyInto(source, target, targetv1beta1.Table_GroupVersionKind, "spec.forProvider.tags"); err != nil {
		return nil, errors.Wrap(err, "failed to copy source into target")
	}
	target.Spec.ForProvider.Tags = make(map[string]*string, len(source.Spec.ForProvider.Tags))
	for _, t := range source.Spec.ForProvider.Tags {
		v := t.Value
		target.Spec.ForProvider.Tags[*t.Key] = v
	}
	target.Spec.ForProvider.Region = &source.Spec.ForProvider.Region
	target.Spec.ForProvider.BillingMode = source.Spec.ForProvider.BillingMode
	for _, t := range source.Spec.ForProvider.AttributeDefinitions {
		parameter := &targetv1beta1.AttributeParameters{Name: t.AttributeName, Type: t.AttributeType}
		target.Spec.ForProvider.Attribute = append(target.Spec.ForProvider.Attribute, *parameter)
	}
	for _, t := range source.Spec.ForProvider.KeySchema {
		if *t.KeyType == "HASH" {
			target.Spec.ForProvider.HashKey = t.AttributeName
		}
		if *t.KeyType == "RANGE" {
			target.Spec.ForProvider.RangeKey = t.AttributeName
		}
	}
	if source.Spec.ForProvider.GlobalSecondaryIndexes != nil {
		for _, t := range source.Spec.ForProvider.GlobalSecondaryIndexes {
			if t.IndexName != nil {
				parameter := &targetv1beta1.GlobalSecondaryIndexParameters{
					Name:             t.IndexName,
					NonKeyAttributes: t.Projection.NonKeyAttributes,
					ProjectionType:   t.Projection.ProjectionType,
				}
				for _, a := range t.KeySchema {
					if *a.KeyType == "HASH" {
						parameter.HashKey = a.AttributeName
					}
					if *a.KeyType == "RANGE" {
						parameter.RangeKey = a.AttributeName
					}
				}
				target.Spec.ForProvider.GlobalSecondaryIndex = append(target.Spec.ForProvider.GlobalSecondaryIndex, *parameter)
			}
		}
	}
	if source.Spec.ForProvider.LocalSecondaryIndexes != nil {
		for _, t := range source.Spec.ForProvider.LocalSecondaryIndexes {
			parameter := &targetv1beta1.LocalSecondaryIndexParameters{
				Name:             t.IndexName,
				NonKeyAttributes: t.Projection.NonKeyAttributes,
				ProjectionType:   t.Projection.ProjectionType,
			}
			for _, a := range t.KeySchema {
				if *a.KeyType == "RANGE" {
					parameter.RangeKey = a.AttributeName
				}
			}
			target.Spec.ForProvider.LocalSecondaryIndex = append(target.Spec.ForProvider.LocalSecondaryIndex, *parameter)
		}
	}

	if source.Spec.ForProvider.ProvisionedThroughput.ReadCapacityUnits != nil {
		convert := float64(*source.Spec.ForProvider.ProvisionedThroughput.ReadCapacityUnits)
		target.Spec.ForProvider.ReadCapacity = &convert
	}
	if source.Spec.ForProvider.ProvisionedThroughput.ReadCapacityUnits != nil {
		convert := float64(*source.Spec.ForProvider.ProvisionedThroughput.ReadCapacityUnits)
		target.Spec.ForProvider.WriteCapacity = &convert
	}
	target.Spec.ForProvider.StreamEnabled = source.Spec.ForProvider.StreamSpecification.StreamEnabled
	target.Spec.ForProvider.StreamViewType = source.Spec.ForProvider.StreamSpecification.StreamViewType
	if source.Spec.ForProvider.SSESpecification != nil {
		target.Spec.ForProvider.ServerSideEncryption = make([]targetv1beta1.ServerSideEncryptionParameters, 1)
		target.Spec.ForProvider.ServerSideEncryption[0].Enabled = source.Spec.ForProvider.SSESpecification.Enabled
		target.Spec.ForProvider.ServerSideEncryption[0].KMSKeyArn = source.Spec.ForProvider.SSESpecification.KMSMasterKeyID
	}

	return []resource.Managed{
		target,
	}, nil
}
