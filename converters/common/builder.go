package common

import (
	"github.com/upbound/upjet/pkg/migration"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

type RegistryBuilder struct {
	registry        *migration.Registry
	converterStores map[schema.GroupVersionKind]ConverterConfig
}

func NewRegistryBuilder(registry *migration.Registry) *RegistryBuilder {
	rb := &RegistryBuilder{
		registry:        registry,
		converterStores: map[schema.GroupVersionKind]ConverterConfig{},
	}
	return rb
}

func (rb *RegistryBuilder) AddConverterStore(gvk schema.GroupVersionKind, cc ConverterConfig) {
	rb.converterStores[gvk] = cc
}

type ConverterConfig struct {
	resourceConverter    migration.ResourceConversionFn
	compositionConverter migration.ComposedTemplateConversionFn
	patchSetConverter    migration.PatchSetsConversionFn
}

func NewConverterConfig(opts ...ConverterStoreOption) ConverterConfig {
	cs := ConverterConfig{}
	for _, o := range opts {
		o(&cs)
	}
	return cs
}

type ConverterStoreOption func(*ConverterConfig)

func WithResourceConverter(resourceConverter migration.ResourceConversionFn) ConverterStoreOption {
	return func(cs *ConverterConfig) {
		cs.resourceConverter = resourceConverter
	}
}

func WithCompositionConverter(compositionConverter migration.ComposedTemplateConversionFn) ConverterStoreOption {
	return func(cs *ConverterConfig) {
		cs.compositionConverter = compositionConverter
	}
}

func WithPatchSetConverter(patchSetConverter migration.PatchSetsConversionFn) ConverterStoreOption {
	return func(cs *ConverterConfig) {
		cs.patchSetConverter = patchSetConverter
	}
}

func (rb *RegistryBuilder) Register() {
	for gvk, converterStore := range rb.converterStores {
		rb.registry.RegisterAPIConversionFunctions(gvk,
			converterStore.resourceConverter,
			converterStore.compositionConverter,
			converterStore.patchSetConverter)
	}
}
