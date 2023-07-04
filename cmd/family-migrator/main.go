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

// main package for the family-migrator tool...
package main

import (
	"k8s.io/apimachinery/pkg/runtime/schema"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/alecthomas/kong"
	"github.com/upbound/upjet/pkg/migration"
	"gopkg.in/yaml.v3"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/upbound/extensions-migration/pkg/converter/configuration"
)

const (
	defaultKubeConfig = ".kube/config"
)

// Options represents the available options for the family-migrator.
type Options struct {
	RegistryOrg                string `name:"regorg" required:"" default:"xpkg.upbound.io/upbound" help:"<registry host>/<organization> for the provider family packages."`
	AWSFamilyVersion           string `name:"aws-family-version" required:"" help:"Version of the AWS provider family."`
	AzureFamilyVersion         string `name:"azure-family-version" required:"" help:"Version of the Azure provider family."`
	GCPFamilyVersion           string `name:"gcp-family-version" required:"" help:"Version of the GCP provider family."`
	SourceConfigurationPackage string `name:"source-configuration-package" required:"" help:"Migration source Configuration package's URL."`
	TargetConfigurationPackage string `name:"target-configuration-package" required:"" help:"Migration target Configuration package's URL."`
	Output                     string `name:"output" required:"" help:"Migration plan output path."`

	PackageRoot   string `name:"package-root" default:"." help:"Source directory for the Crossplane Configuration package."`
	ExamplesRoot  string `name:"examples-root" default:"./examples" help:"Path to Crossplane package examples directory."`
	PackageOutput string `name:"package-output" default:"updated-configuration.pkg" help:"Path to store the updated configuration package."`

	KubeConfig string `name:"kubeconfig" optional:"" help:"Path to the kubeconfig to use."`
}

func main() {
	opts := &Options{}
	kongCtx := kong.Parse(opts, kong.Name("family-migrator"),
		kong.Description("Upbound provider families migration tool"),
		kong.UsageOnError(),
		kong.ConfigureHelp(kong.HelpOptions{
			Compact:   true,
			FlagsLast: true,
			Summary:   true,
		}))

	r := migration.NewRegistry(runtime.NewScheme())
	err := r.AddCrossplanePackageTypes()
	kongCtx.FatalIfErrorf(err, "Failed to register the Provider package types with the migration registry")
	cp := configuration.NewCompositionPreProcessor()
	r.RegisterPreProcessor(migration.CategoryComposition, migration.PreProcessor(cp.GetSSOPNameFromComposition))
	r.RegisterConfigurationMetadataConverter(migration.AllConfigurations, &configuration.ConfigMetaParameters{
		FamilyVersion:        opts.AWSFamilyVersion,
		Monolith:             "provider-aws",
		CompositionProcessor: cp,
	})
	r.RegisterConfigurationMetadataConverter(migration.AllConfigurations, &configuration.ConfigMetaParameters{
		FamilyVersion:        opts.AzureFamilyVersion,
		Monolith:             "provider-azure",
		CompositionProcessor: cp,
	})
	r.RegisterConfigurationMetadataConverter(migration.AllConfigurations, &configuration.ConfigMetaParameters{
		FamilyVersion:        opts.GCPFamilyVersion,
		Monolith:             "provider-gcp",
		CompositionProcessor: cp,
	})
	r.RegisterConfigurationPackageConverter(regexp.MustCompile(opts.SourceConfigurationPackage), &configuration.ConfigPkgParameters{
		PackageURL: opts.TargetConfigurationPackage,
	})
	// TODO: should we also handle missing registry (xpkg.upbound.io),
	// i.e., is it the default?
	// register converters for the family config packages
	r.RegisterProviderPackageConverter(regexp.MustCompile(`xpkg.upbound.io/upbound/provider-aws:.+`), &configuration.ProviderPkgFamilyConfigParameters{
		FamilyVersion: opts.AWSFamilyVersion,
	})
	r.RegisterProviderPackageConverter(regexp.MustCompile(`xpkg.upbound.io/upbound/provider-azure:.+`), &configuration.ProviderPkgFamilyConfigParameters{
		FamilyVersion: opts.AzureFamilyVersion,
	})
	r.RegisterProviderPackageConverter(regexp.MustCompile(`xpkg.upbound.io/upbound/provider-gcp:.+`), &configuration.ProviderPkgFamilyConfigParameters{
		FamilyVersion: opts.GCPFamilyVersion,
	})
	// register converters for the family resource packages
	r.RegisterProviderPackageConverter(regexp.MustCompile(`xpkg.upbound.io/upbound/provider-aws:.+`), &configuration.ProviderPkgFamilyParameters{
		FamilyVersion:        opts.AWSFamilyVersion,
		Monolith:             "provider-aws",
		CompositionProcessor: cp,
	})
	r.RegisterProviderPackageConverter(regexp.MustCompile(`xpkg.upbound.io/upbound/provider-azure:.+`), &configuration.ProviderPkgFamilyParameters{
		FamilyVersion:        opts.AzureFamilyVersion,
		Monolith:             "provider-azure",
		CompositionProcessor: cp,
	})
	r.RegisterProviderPackageConverter(regexp.MustCompile(`xpkg.upbound.io/upbound/provider-gcp:.+`), &configuration.ProviderPkgFamilyParameters{
		FamilyVersion:        opts.GCPFamilyVersion,
		Monolith:             "provider-gcp",
		CompositionProcessor: cp,
	})
	r.RegisterPackageLockConverter(migration.CrossplaneLockName, &configuration.LockParameters{
		PackageURL: opts.SourceConfigurationPackage,
	})
	kongCtx.FatalIfErrorf(r.AddCompositionTypes(), "Failed to register the Crossplane Composition types with the migration registry")

	fsSource, err := migration.NewFileSystemSource(opts.PackageRoot)
	kongCtx.FatalIfErrorf(err, "Failed to initialize the migration FileSystem source from path: %s", opts.PackageRoot)

	if len(opts.KubeConfig) == 0 {
		homeDir, err := os.UserHomeDir()
		kongCtx.FatalIfErrorf(err, "Failed to get user's home")
		opts.KubeConfig = filepath.Join(homeDir, defaultKubeConfig)
	}
	kubeSource, err := migration.NewKubernetesSourceFromKubeConfig(opts.KubeConfig, migration.WithRegistry(r) /*migration.WithCategories([]migration.Category{migration.CategoryManaged})*/)
	kongCtx.FatalIfErrorf(err, "Failed to initialize the migration Kubernetes source from kubeconfig: %s", opts.KubeConfig)
	pg := migration.NewPlanGenerator(r, nil, migration.NewFileSystemTarget(), migration.WithEnableConfigurationMigrationSteps(), migration.WithMultipleSources(fsSource, kubeSource), migration.WithSkipGVKs(schema.GroupVersionKind{}))
	kongCtx.FatalIfErrorf(pg.GeneratePlan(), "Failed to generate the migration plan for the provider families")
	setPkgParameters(&pg.Plan, *opts)
	buff, err := yaml.Marshal(pg.Plan)
	kongCtx.FatalIfErrorf(err, "Failed to marshal the migration plan to YAML")
	kongCtx.FatalIfErrorf(os.WriteFile(opts.Output, buff, 0600), "Failed to store the migration plan at path: %s", opts.Output)
}

func setPkgParameters(plan *migration.Plan, opts Options) {
	for i, s := range plan.Spec.Steps {
		// TODO: consider exporting step constants. But the idea is
		// to introduce the concept of a migration Scenario that
		// encapsulated both the converters and the steps involved.
		if s.Name == "push-configuration" || s.Name == "build-configuration" {
			s.Exec.Args[1] = strings.ReplaceAll(s.Exec.Args[1], "{{TARGET_CONFIGURATION_PACKAGE}}", opts.TargetConfigurationPackage)
			s.Exec.Args[1] = strings.ReplaceAll(s.Exec.Args[1], "{{PKG_PATH}}", opts.PackageOutput)
			s.Exec.Args[1] = strings.ReplaceAll(s.Exec.Args[1], "{{PKG_ROOT}}", opts.PackageRoot)
			s.Exec.Args[1] = strings.ReplaceAll(s.Exec.Args[1], "{{EXAMPLES_ROOT}}", opts.ExamplesRoot)
			migration.AddExecStepManualExecution(&s)
			plan.Spec.Steps[i] = s
		}
	}
}
