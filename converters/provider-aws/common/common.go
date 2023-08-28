package common

import (
	"fmt"
	"strings"

	"github.com/crossplane/crossplane-runtime/pkg/fieldpath"
	xpv1 "github.com/crossplane/crossplane/apis/apiextensions/v1"
	"github.com/pkg/errors"
	"github.com/upbound/upjet/pkg/migration"
)

// DefaultPatchSetsConverter is a default patchset converter for the community provider-aws
func DefaultPatchSetsConverter(sourcePatchSets map[string]*xpv1.PatchSet) error {
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

	tPs := sourcePatchSets[tagsPatchSetName]
	if tPs == nil {
		return nil
	}
	for i, p := range tPs.Patches {
		r := strings.NewReplacer("metadata.labels[", "", "]", "")
		key := r.Replace(*p.FromFieldPath)
		*tPs.Patches[i].ToFieldPath = fmt.Sprintf(`spec.forProvider.tags[%s]`, key)
	}
	// convert patch sets in the source
	return nil
}

// ConvertComposedTemplateTags is responsible converting the tags of provider-aws resources
func ConvertComposedTemplateTags(sourceTemplate xpv1.ComposedTemplate) ([]xpv1.Patch, error) {
	var patchesToAdd []xpv1.Patch
	for _, p := range sourceTemplate.Patches {
		if p.ToFieldPath != nil {
			if strings.HasPrefix(*p.ToFieldPath, "spec.forProvider.tags") {
				u, err := migration.FromRawExtension(sourceTemplate.Base)
				if err != nil {
					return nil, errors.Wrap(err, "failed to convert ComposedTemplate")
				}
				paved := fieldpath.Pave(u.Object)
				key, err := paved.GetString(strings.ReplaceAll(*p.ToFieldPath, ".value", ".key"))
				if err != nil {
					return nil, errors.Wrap(err, "failed to get value from paved")
				}
				s := fmt.Sprintf(`spec.forProvider.tags["%s"]`, key)
				patchesToAdd = append(patchesToAdd, xpv1.Patch{
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
