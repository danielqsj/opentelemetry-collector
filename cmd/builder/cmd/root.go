// Copyright The OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd // import "go.opentelemetry.io/collector/cmd/builder/cmd"

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap"

	"go.opentelemetry.io/collector/cmd/builder/internal/builder"
)

var (
	version = "dev"
	date    = "unknown"

	cfgFile string
	cfg     = builder.DefaultConfig()

	versionCmd = &cobra.Command{
		Use:   "version",
		Short: "Version of opentelemetry-collector-builder",
		Long:  "Prints the version of opentelemetry-collector-builder binary",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Println(fmt.Sprintf("%s version %s", cmd.Parent().Name(), version))
		},
	}
)

// Execute is the main entrypoint for this application
func Execute() error {
	cobra.OnInitialize(initConfig)

	cmd := &cobra.Command{
		Use:  "builder",
		Long: fmt.Sprintf("OpenTelemetry Collector distribution builder (%s)", version),
		RunE: func(cmd *cobra.Command, args []string) error {

			if err := cfg.Validate(); err != nil {
				cfg.Logger.Error("invalid configuration", zap.Error(err))
				return err
			}

			if err := cfg.ParseModules(); err != nil {
				cfg.Logger.Error("invalid module configuration", zap.Error(err))
				return err
			}

			return builder.GenerateAndCompile(cfg)
		},
	}

	// the external config file
	cmd.Flags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.otelcol-builder.yaml)")

	// the distribution parameters, which we accept as CLI flags as well
	cmd.Flags().BoolVar(&cfg.SkipCompilation, "skip-compilation", false, "Whether builder should only generate go code with no compile of the collector (default false)")
	cmd.Flags().StringVar(&cfg.Distribution.ExeName, "name", "otelcol-custom", "The executable name for the OpenTelemetry Collector distribution")
	cmd.Flags().StringVar(&cfg.Distribution.LongName, "description", "Custom OpenTelemetry Collector distribution", "A descriptive name for the OpenTelemetry Collector distribution")
	cmd.Flags().StringVar(&cfg.Distribution.Version, "version", "1.0.0", "The version for the OpenTelemetry Collector distribution")
	cmd.Flags().BoolVar(&cfg.Distribution.IncludeCore, "include-core", true, "Whether the core components should be included in the distribution")
	cmd.Flags().StringVar(&cfg.Distribution.OtelColVersion, "otelcol-version", cfg.Distribution.OtelColVersion, "Which version of OpenTelemetry Collector to use as base")
	cmd.Flags().StringVar(&cfg.Distribution.OutputPath, "output-path", cfg.Distribution.OutputPath, "Where to write the resulting files")
	cmd.Flags().StringVar(&cfg.Distribution.Go, "go", "", "The Go binary to use during the compilation phase. Default: go from the PATH")
	cmd.Flags().StringVar(&cfg.Distribution.Module, "module", "go.opentelemetry.io/collector/cmd/builder", "The Go module for the new distribution")

	// version of this binary
	cmd.AddCommand(versionCmd)

	// tie Viper to flags
	if err := viper.BindPFlags(cmd.Flags()); err != nil {
		cfg.Logger.Error("failed to bind flags", zap.Error(err))
		return err
	}

	if err := cmd.Execute(); err != nil {
		cfg.Logger.Error("failed to run", zap.Error(err))
		return err
	}

	return nil
}

func initConfig() {
	cfg.Logger.Info("OpenTelemetry Collector distribution builder", zap.String("version", version), zap.String("date", date))

	vp := viper.New()

	// a couple of Viper goodies, to make it easier to use env vars when flags are not desirable
	vp.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	vp.AutomaticEnv()

	// load values from config file -- required for the modules configuration
	if cfgFile != "" {
		vp.SetConfigFile(cfgFile)
	} else {
		vp.AddConfigPath("$HOME")
		vp.SetConfigName(".otelcol-builder")
	}

	// load the config file
	if err := vp.ReadInConfig(); err != nil {
		cobra.CheckErr(err)
	}
	cfg.Logger.Info("Using config file", zap.String("path", vp.ConfigFileUsed()))

	// convert Viper's internal state into our configuration object
	if err := vp.Unmarshal(&cfg); err != nil {
		cfg.Logger.Error("failed to parse the config", zap.Error(err))
		cobra.CheckErr(err)
		return
	}
}
