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

package sqs

import (
	"encoding/json"

	srcv1alpha1 "github.com/crossplane-contrib/provider-aws/apis/sqs/v1beta1"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/crossplane/upjet/pkg/migration"
	"github.com/pkg/errors"
	targetv1beta1 "github.com/upbound/provider-aws/apis/sqs/v1beta1"
)

func QueueResource(mg resource.Managed) ([]resource.Managed, error) {
	source := mg.(*srcv1alpha1.Queue)
	target := &targetv1beta1.Queue{}
	if _, err := migration.CopyInto(source, target, targetv1beta1.Queue_GroupVersionKind, "spec.forProvider.redrivePolicy"); err != nil {
		return nil, errors.Wrap(err, "failed to copy source into target")
	}

	target.Spec.ForProvider.Region = &source.Spec.ForProvider.Region
	if source.Spec.ForProvider.DelaySeconds != nil {
		convert := float64(*source.Spec.ForProvider.DelaySeconds)
		target.Spec.ForProvider.DelaySeconds = &convert
	}
	if source.Spec.ForProvider.MaximumMessageSize != nil {
		convert := float64(*source.Spec.ForProvider.MaximumMessageSize)
		target.Spec.ForProvider.MaxMessageSize = &convert
	}
	if source.Spec.ForProvider.MessageRetentionPeriod != nil {
		convert := float64(*source.Spec.ForProvider.MessageRetentionPeriod)
		target.Spec.ForProvider.MessageRetentionSeconds = &convert
	}
	if source.Spec.ForProvider.ReceiveMessageWaitTimeSeconds != nil {
		convert := float64(*source.Spec.ForProvider.ReceiveMessageWaitTimeSeconds)
		target.Spec.ForProvider.ReceiveWaitTimeSeconds = &convert
	}
	if source.Spec.ForProvider.VisibilityTimeout != nil {
		convert := float64(*source.Spec.ForProvider.VisibilityTimeout)
		target.Spec.ForProvider.VisibilityTimeoutSeconds = &convert
	}
	target.Spec.ForProvider.SqsManagedSseEnabled = source.Spec.ForProvider.SqsManagedSseEnabled
	target.Spec.ForProvider.KMSMasterKeyID = source.Spec.ForProvider.KMSMasterKeyID
	target.Spec.ForProvider.Policy = source.Spec.ForProvider.Policy
	target.Spec.ForProvider.FifoQueue = source.Spec.ForProvider.FIFOQueue
	target.Spec.ForProvider.ContentBasedDeduplication = source.Spec.ForProvider.ContentBasedDeduplication

	if source.Spec.ForProvider.RedrivePolicy != nil {
		RedrivePolicyData := map[string]interface{}{
			"deadLetterTargetArn": source.Spec.ForProvider.RedrivePolicy.DeadLetterTargetARN,
			"maxReceiveCount":     source.Spec.ForProvider.RedrivePolicy.MaxReceiveCount,
		}

		RedrivePolicyDataJson, err := json.Marshal(RedrivePolicyData)
		if err != nil {
			return nil, errors.Wrap(err, "failed to parse redrive policy")
		}
		convert := string(RedrivePolicyDataJson)
		target.Spec.ForProvider.RedrivePolicy = &convert
	}

	return []resource.Managed{
		target,
	}, nil
}
