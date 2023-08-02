package null

import (
	"github.com/upbound/extensions-migration/converters/common"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func ExampleMRKindConfigurator(builder *common.RegistryBuilder) {
	builder.AddConverterStore(schema.GroupVersionKind{},
		common.NewConverterConfig(
		// Add the required converters to the corresponding functions.
		//
		// common.WithResourceConverter( * Resource Converter * ),
		// common.WithCompositionConverter( * Composition Converter * ),
		// common.WithPatchSetConverter( * PatchSet Converter * ),
		))
}
