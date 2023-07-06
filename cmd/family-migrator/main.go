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
	"github.com/crossplane/crossplane-runtime/pkg/logging"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"os"
	"path/filepath"
	"regexp"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
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
	Generate struct {
		RegistryOrg                string `name:"regorg" required:"" default:"xpkg.upbound.io/upbound" help:"<registry host>/<organization> for the provider family packages."`
		AWSFamilyVersion           string `name:"aws-family-version" required:"" help:"Version of the AWS provider family."`
		AzureFamilyVersion         string `name:"azure-family-version" required:"" help:"Version of the Azure provider family."`
		GCPFamilyVersion           string `name:"gcp-family-version" required:"" help:"Version of the GCP provider family."`
		SourceConfigurationPackage string `name:"source-configuration-package" required:"" help:"Migration source Configuration package's URL."`
		TargetConfigurationPackage string `name:"target-configuration-package" required:"" help:"Migration target Configuration package's URL."`

		PackageRoot   string `name:"package-root" default:"package" help:"Source directory for the Crossplane Configuration package."`
		ExamplesRoot  string `name:"examples-root" default:"package/examples" help:"Path to Crossplane package examples directory."`
		PackageOutput string `name:"package-output" default:"updated-configuration.pkg" help:"Path to store the updated configuration package."`

		KubeConfig string `name:"kubeconfig" optional:"" help:"Path to the kubeconfig to use."`
	} `kong:"cmd"`

	Execute struct{} `kong:"cmd"`

	PlanPath string `name:"plan-path" default:"migration-plan.yaml" help:"Migration plan output path."`

	Debug bool `name:"debug" short:"d" optional:"" help:"Run with debug logging."`
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

	absPath, err := filepath.Abs(opts.PlanPath)
	kongCtx.FatalIfErrorf(err, "Failed to get the absolute path for the migration plan output: %s", opts.PlanPath)
	planDir := filepath.Dir(absPath)

	switch kongCtx.Command() {
	case "generate":
		generatePlan(kongCtx, opts, planDir)
	case "execute":
		executePlan(kongCtx, planDir, opts)
	}
}

func generatePlan(kongCtx *kong.Context, opts *Options, planDir string) {
	r := migration.NewRegistry(runtime.NewScheme())
	err := r.AddCrossplanePackageTypes()
	kongCtx.FatalIfErrorf(err, "Failed to register the Provider package types with the migration registry")
	cp := configuration.NewCompositionPreProcessor()
	r.RegisterPreProcessor(migration.CategoryComposition, migration.PreProcessor(cp.GetSSOPNameFromComposition))
	r.RegisterConfigurationMetadataConverter(migration.AllConfigurations, &configuration.ConfigMetaParameters{
		FamilyVersion:        opts.Generate.AWSFamilyVersion,
		Monolith:             "provider-aws",
		CompositionProcessor: cp,
	})
	r.RegisterConfigurationMetadataConverter(migration.AllConfigurations, &configuration.ConfigMetaParameters{
		FamilyVersion:        opts.Generate.AzureFamilyVersion,
		Monolith:             "provider-azure",
		CompositionProcessor: cp,
	})
	r.RegisterConfigurationMetadataConverter(migration.AllConfigurations, &configuration.ConfigMetaParameters{
		FamilyVersion:        opts.Generate.GCPFamilyVersion,
		Monolith:             "provider-gcp",
		CompositionProcessor: cp,
	})
	r.RegisterConfigurationPackageConverter(regexp.MustCompile(opts.Generate.SourceConfigurationPackage), &configuration.ConfigPkgParameters{
		PackageURL: opts.Generate.TargetConfigurationPackage,
	})
	// TODO: should we also handle missing registry (xpkg.upbound.io),
	// i.e., is it the default?
	// register converters for the family config packages
	r.RegisterProviderPackageConverter(regexp.MustCompile(`xpkg.upbound.io/upbound/provider-aws:.+`), &configuration.ProviderPkgFamilyConfigParameters{
		FamilyVersion: opts.Generate.AWSFamilyVersion,
	})
	r.RegisterProviderPackageConverter(regexp.MustCompile(`xpkg.upbound.io/upbound/provider-azure:.+`), &configuration.ProviderPkgFamilyConfigParameters{
		FamilyVersion: opts.Generate.AzureFamilyVersion,
	})
	r.RegisterProviderPackageConverter(regexp.MustCompile(`xpkg.upbound.io/upbound/provider-gcp:.+`), &configuration.ProviderPkgFamilyConfigParameters{
		FamilyVersion: opts.Generate.GCPFamilyVersion,
	})
	// register converters for the family resource packages
	r.RegisterProviderPackageConverter(regexp.MustCompile(`xpkg.upbound.io/upbound/provider-aws:.+`), &configuration.ProviderPkgFamilyParameters{
		FamilyVersion:        opts.Generate.AWSFamilyVersion,
		Monolith:             "provider-aws",
		CompositionProcessor: cp,
	})
	r.RegisterProviderPackageConverter(regexp.MustCompile(`xpkg.upbound.io/upbound/provider-azure:.+`), &configuration.ProviderPkgFamilyParameters{
		FamilyVersion:        opts.Generate.AzureFamilyVersion,
		Monolith:             "provider-azure",
		CompositionProcessor: cp,
	})
	r.RegisterProviderPackageConverter(regexp.MustCompile(`xpkg.upbound.io/upbound/provider-gcp:.+`), &configuration.ProviderPkgFamilyParameters{
		FamilyVersion:        opts.Generate.GCPFamilyVersion,
		Monolith:             "provider-gcp",
		CompositionProcessor: cp,
	})
	r.RegisterPackageLockConverter(migration.CrossplaneLockName, &configuration.LockParameters{
		PackageURL: opts.Generate.SourceConfigurationPackage,
	})
	kongCtx.FatalIfErrorf(r.AddCompositionTypes(), "Failed to register the Crossplane Composition types with the migration registry")

	fsSource, err := migration.NewFileSystemSource(opts.Generate.PackageRoot)
	kongCtx.FatalIfErrorf(err, "Failed to initialize the migration FileSystem source from path: %s", opts.Generate.PackageRoot)

	if len(opts.Generate.KubeConfig) == 0 {
		homeDir, err := os.UserHomeDir()
		kongCtx.FatalIfErrorf(err, "Failed to get user's home")
		opts.Generate.KubeConfig = filepath.Join(homeDir, defaultKubeConfig)
	}
	kubeSource, err := migration.NewKubernetesSourceFromKubeConfig(opts.Generate.KubeConfig, migration.WithRegistry(r), migration.WithCategories([]migration.Category{migration.CategoryManaged}))
	kongCtx.FatalIfErrorf(err, "Failed to initialize the migration Kubernetes source from kubeconfig: %s", opts.Generate.KubeConfig)

	pg := migration.NewPlanGenerator(r, nil, migration.NewFileSystemTarget(migration.WithParentDirectory(planDir)), migration.WithEnableConfigurationMigrationSteps(), migration.WithMultipleSources(fsSource, kubeSource), migration.WithSkipGVKs(schema.GroupVersionKind{}))
	kongCtx.FatalIfErrorf(pg.GeneratePlan(), "Failed to generate the migration plan for the provider families")

	setPkgParameters(&pg.Plan, *opts)
	buff, err := yaml.Marshal(pg.Plan)
	kongCtx.FatalIfErrorf(err, "Failed to marshal the migration plan to YAML")
	kongCtx.FatalIfErrorf(os.WriteFile(opts.PlanPath, buff, 0600), "Failed to store the migration plan at path: %s", opts.PlanPath)
}

func executePlan(kongCtx *kong.Context, planDir string, opts *Options) {
	plan := &migration.Plan{}
	buff, err := os.ReadFile(opts.PlanPath)
	kongCtx.FatalIfErrorf(err, "Failed to read the migration plan from path: %s", opts.PlanPath)
	kongCtx.FatalIfErrorf(yaml.Unmarshal(buff, plan), "Failed to unmarshal the migration plan: %s", opts.PlanPath)
	zl := zap.New(zap.UseDevMode(opts.Debug))
	log := logging.NewLogrLogger(zl.WithName("fork-executor"))
	executor := migration.NewForkExecutor(migration.WithWorkingDir(planDir), migration.WithLogger(log))
	// TODO: we need to load the plan back from the filesystem as it may
	// have been modified.
	planExecutor := migration.NewPlanExecutor(*plan, []migration.Executor{executor},
		migration.WithExecutorCallback(&executionCallback{
			logger: logging.NewLogrLogger(zl.WithName("family-migrator")),
		}))
	backupDir := filepath.Join(planDir, "backup")
	kongCtx.FatalIfErrorf(os.MkdirAll(backupDir, 0o700), "Failed to mkdir backup directory: %s", backupDir)
	kongCtx.FatalIfErrorf(planExecutor.Execute(), "Failed to execute the migration plan at path: %s", opts.PlanPath)
}

func setPkgParameters(plan *migration.Plan, opts Options) {
	for i, s := range plan.Spec.Steps {
		// TODO: consider exporting step constants. But the idea is
		// to introduce the concept of a migration Scenario that
		// encapsulated both the converters and the steps involved.
		if s.Name == "push-configuration" || s.Name == "build-configuration" {
			s.Exec.Args[1] = strings.ReplaceAll(s.Exec.Args[1], "{{TARGET_CONFIGURATION_PACKAGE}}", opts.Generate.TargetConfigurationPackage)
			s.Exec.Args[1] = strings.ReplaceAll(s.Exec.Args[1], "{{PKG_PATH}}", opts.Generate.PackageOutput)
			s.Exec.Args[1] = strings.ReplaceAll(s.Exec.Args[1], "{{PKG_ROOT}}", opts.Generate.PackageRoot)
			s.Exec.Args[1] = strings.ReplaceAll(s.Exec.Args[1], "{{EXAMPLES_ROOT}}", opts.Generate.ExamplesRoot)
			migration.AddManualExecution(&s)
			plan.Spec.Steps[i] = s
		}
	}
}

type executionCallback struct {
	logger logging.Logger
}

func (cb *executionCallback) StepToExecute(s migration.Step, index int) migration.CallbackResult {
	cb.logger.Info("Executing step...", "index", index, "name", s.Name)
	return migration.CallbackResult{Action: migration.ActionContinue}
}

func (cb *executionCallback) StepSucceeded(s migration.Step, index int, buff []byte) migration.CallbackResult {
	cb.logger.Info("Step succeeded", "output", string(buff), "index", index, "name", s.Name)
	return migration.CallbackResult{Action: migration.ActionContinue}
}

func (cb *executionCallback) StepFailed(s migration.Step, index int, buff []byte, err error) migration.CallbackResult {
	cb.logger.Info("Step failed", "output", string(buff), "index", index, "name", s.Name, "err", err)
	return migration.CallbackResult{Action: migration.ActionCancel}
}
