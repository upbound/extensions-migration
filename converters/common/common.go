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

package common

import (
	"fmt"
	"k8s.io/apimachinery/pkg/runtime"
	"strings"

	"github.com/crossplane/crossplane-runtime/pkg/fieldpath"
	v1 "github.com/crossplane/crossplane/apis/apiextensions/v1"
	"github.com/pkg/errors"
	"github.com/upbound/upjet/pkg/migration"
)

func AddToScheme(r *migration.Registry, sourceF, targetF func(s *runtime.Scheme) error) error {
	if err := r.AddToScheme(sourceF); err != nil {
		return err
	}
	if err := r.AddToScheme(targetF); err != nil {
		return err
	}
	return nil
}

func PtrFromString(s string) *string {
	return &s
}

func PtrFloat64(i *int32) *float64 {
	a := float64(*i)
	return &a
}

func ConvertTags(sourceTemplate v1.ComposedTemplate, value string, key string) ([]v1.Patch, error) {
	patchesToAdd := []v1.Patch{}
	for _, p := range sourceTemplate.Patches {
		if p.ToFieldPath != nil {
			if strings.HasPrefix(*p.ToFieldPath, "spec.forProvider.tags") {
				u, err := migration.FromRawExtension(sourceTemplate.Base)
				if err != nil {
					return nil, errors.Wrap(err, "failed to convert ComposedTemplate")
				}
				paved := fieldpath.Pave(u.Object)
				key, err := paved.GetString(strings.ReplaceAll(*p.ToFieldPath, value, key))
				if err != nil {
					return nil, errors.Wrap(err, "failed to get value from paved")
				}
				s := fmt.Sprintf(`spec.forProvider.tags["%s"]`, key)
				patchesToAdd = append(patchesToAdd, v1.Patch{
					FromFieldPath: p.FromFieldPath,
					ToFieldPath:   &s,
					Transforms:    p.Transforms,
					Policy:        p.Policy,
				})
			}
		}
	}
	return patchesToAdd, nil
}

func DefaultCompositionConverter(convertTags bool, conversionMap map[string]string, sourceTemplate v1.ComposedTemplate, convertedTemplates ...*v1.ComposedTemplate) error {
	var patchesToAdd []v1.Patch
	var err error
	if convertTags {
		patchesToAdd, err = ConvertTags(sourceTemplate, ".value", ".key")
		if err != nil {
			return errors.Wrap(err, "failed to convert tags")
		}
	}
	patchesToAdd = append(patchesToAdd, ConvertPatchesMap(sourceTemplate, conversionMap)...)
	for i := range convertedTemplates {
		convertedTemplates[i].Patches = append(convertedTemplates[i].Patches, patchesToAdd...)
	}
	return nil
}

func getTagsPatchSetName(sourcePatchSets map[string]*v1.PatchSet) string {
	tagsPatchSetName := ""
	for _, patchSet := range sourcePatchSets {
		for _, patch := range patchSet.Patches {
			if patch.ToFieldPath != nil {
				if strings.HasPrefix(*patch.ToFieldPath, "spec.forProvider.tags") {
					tagsPatchSetName = patchSet.Name
					break
				}
			}
		}
		if tagsPatchSetName != "" {
			break
		}
	}
	return tagsPatchSetName
}

func convertTagsInPatchset(psMap map[string]*v1.PatchSet, patchSetName string) error {

	tPs := psMap[patchSetName]
	if tPs == nil {
		return nil
	}
	for i, p := range tPs.Patches {
		r := strings.NewReplacer("metadata.labels[", "", "]", "")
		key := r.Replace(*p.FromFieldPath)
		*tPs.Patches[i].ToFieldPath = fmt.Sprintf(`spec.forProvider.tags[%s]`, key)
	}
	return nil
}

func ConvertPatchSets(sourcePatchSets map[string]*v1.PatchSet) error {
	tagsPatchSetName := getTagsPatchSetName(sourcePatchSets)
	// convert patch sets in the source
	return errors.Wrap(convertTagsInPatchset(sourcePatchSets, tagsPatchSetName), "failed to convert patch sets")
}

func ConvertPatchesMap(sourceTemplate v1.ComposedTemplate, conversionMap map[string]string) []v1.Patch {
	patchesToAdd := []v1.Patch{}
	for _, p := range sourceTemplate.Patches {
		switch p.Type {
		case v1.PatchTypeFromCompositeFieldPath, v1.PatchTypeCombineFromComposite, "":
			{
				if p.ToFieldPath != nil {
					if to, ok := conversionMap[*p.ToFieldPath]; ok {
						patchesToAdd = append(patchesToAdd, v1.Patch{
							Type:          p.Type,
							FromFieldPath: p.FromFieldPath,
							ToFieldPath:   &to,
							Transforms:    p.Transforms,
							Policy:        p.Policy,
							Combine:       p.Combine,
						})
					}
				}
			}
		case v1.PatchTypeToCompositeFieldPath, v1.PatchTypeCombineToComposite:
			{
				if p.FromFieldPath != nil {
					if to, ok := conversionMap[*p.FromFieldPath]; ok {
						patchesToAdd = append(patchesToAdd, v1.Patch{
							Type:          p.Type,
							FromFieldPath: &to,
							ToFieldPath:   p.ToFieldPath,
							Transforms:    p.Transforms,
							Policy:        p.Policy,
							Combine:       p.Combine,
						})
					}
				}
			}
		}
	}
	return patchesToAdd
}

func SplittedResourcePatches(convertedTemplates []*v1.ComposedTemplate, resourceName string, patchesToAdd []v1.Patch) error {
	for i, cb := range convertedTemplates {
		if cb.Base.Raw != nil {
			u, err := migration.FromRawExtension(cb.Base)
			if err != nil {
				return errors.Wrap(err, "failed to convert ComposedTemplate base")
			}
			if u.GetKind() == resourceName {
				convertedTemplates[i].Patches = append(convertedTemplates[i].Patches, patchesToAdd...)
			}
		}
	}
	return nil
}
