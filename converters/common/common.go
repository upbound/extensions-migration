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

package common

import (
	v1 "github.com/crossplane/crossplane/apis/apiextensions/v1"
	"github.com/pkg/errors"
	"github.com/upbound/upjet/pkg/migration"
	"k8s.io/apimachinery/pkg/runtime"
)

// AddToScheme adds source and target schemes via registered functions
func AddToScheme(r *migration.Registry, sourceF, targetF func(s *runtime.Scheme) error) error {
	if err := r.AddToScheme(sourceF); err != nil {
		return err
	}
	if err := r.AddToScheme(targetF); err != nil {
		return err
	}
	return nil
}

// PtrFromString returns the parameter of the type string in the pointer type.
func PtrFromString(s string) *string {
	return &s
}

// PtrFloat64FromInt32 returns the parameter of the type int32 in the pointer type.
func PtrFloat64FromInt32(i *int32) *float64 {
	if i == nil {
		return nil
	}
	a := float64(*i)
	return &a
}

// PtrFloat64FromInt64 returns the parameter of the type int64 in the pointer type.
func PtrFloat64FromInt64(i *int64) *float64 {
	if i == nil {
		return nil
	}
	a := float64(*i)
	return &a
}

// SplittedResourcePatches is used when more than one target is generated from the source ComposedTemplate.
// The patch statements are classified for the correct type of resources.
func SplittedResourcePatches(convertedTemplates []*v1.ComposedTemplate, resourceKind string, patchesToAdd []v1.Patch) error {
	for i, cb := range convertedTemplates {
		if cb.Base.Raw != nil {
			u, err := migration.FromRawExtension(cb.Base)
			if err != nil {
				if cb.Name != nil {
					return errors.Wrapf(err, "failed to convert ComposedTemplate base: %s", *cb.Name)
				}
				return errors.Wrapf(err, "failed to convert ComposedTemplate base")
			}
			if u.GetKind() == resourceKind {
				convertedTemplates[i].Patches = append(convertedTemplates[i].Patches, patchesToAdd...)
			}
		}
	}
	return nil
}
