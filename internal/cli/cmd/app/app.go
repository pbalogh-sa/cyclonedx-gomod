// This file is part of CycloneDX GoMod
//
// Licensed under the Apache License, Version 2.0 (the “License”);
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an “AS IS” BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0
// Copyright (c) OWASP Foundation. All Rights Reserved.

package app

import (
	"context"
	"flag"
	"fmt"

	"github.com/peterbourgon/ff/v3/ffcli"

	cliUtil "github.com/CycloneDX/cyclonedx-gomod/internal/cli/util"
	"github.com/CycloneDX/cyclonedx-gomod/internal/sbom"
	"github.com/CycloneDX/cyclonedx-gomod/pkg/generate/app"
	"github.com/CycloneDX/cyclonedx-gomod/pkg/licensedetect"
	"github.com/CycloneDX/cyclonedx-gomod/pkg/licensedetect/local"
)

func New() *ffcli.Command {
	fs := flag.NewFlagSet("cyclonedx-gomod app", flag.ExitOnError)

	var options Options
	options.RegisterFlags(fs)

	return &ffcli.Command{
		Name:       "app",
		ShortHelp:  "Generate SBOMs for applications",
		ShortUsage: "cyclonedx-gomod app [FLAGS...] [MODULE_PATH]",
		LongHelp: `Generate SBOMs for applications.

In order to produce accurate SBOMs, build constraints must be configured
via environment variables. These build constraints should mimic the ones passed
to the "go build" command for the application.

Environment variables that act as build constraints are:
  - GOARCH       The target architecture (386, amd64, etc.)
  - GOOS         The target operating system (linux, windows, etc.)
  - CGO_ENABLED  Whether or not CGO is enabled
  - GOFLAGS      Flags that are passed to the Go command (e.g. build tags)

A complete overview of all environment variables can be found here:
  https://pkg.go.dev/cmd/go#hdr-Environment_variables

Applicable build constraints are included as properties of the main component.

Because build constraints influence Go's module selection, an SBOM should be generated
for each target in the build matrix.

The -main flag should be used to specify the path to the application's main package.
It must point to a directory within MODULE_PATH. If not set, MODULE_PATH is assumed.

In order to not only include modules, but also the packages within them,
the -packages flag can be used. Packages are represented as subcomponents of modules.

By passing -files, all files that would be included in a binary will be attached
as subcomponents of their respective package. File versions follow the v0.0.0-SHORTHASH pattern, 
where SHORTHASH is the first 12 characters of the file's SHA1 hash.
Because files are subcomponents of packages, -files can only be used in conjunction with -packages.

Examples:
  $ GOARCH=arm64 GOOS=linux GOFLAGS="-tags=foo,bar" cyclonedx-gomod app -output linux-arm64.bom.xml
  $ cyclonedx-gomod app -json -output acme-app.bom.json -files -licenses -main cmd/acme-app /usr/src/acme-module`,
		FlagSet: fs,
		Exec: func(_ context.Context, args []string) error {
			if len(args) > 1 {
				return fmt.Errorf("too many arguments (expected 1, got %d)", len(args))
			}
			if len(args) == 0 {
				options.ModuleDir = "."
			} else {
				options.ModuleDir = args[0]
			}

			return Exec(options)
		},
	}
}

func Exec(options Options) error {
	err := options.Validate()
	if err != nil {
		return err
	}

	logger := options.Logger()

	var licenseDetector licensedetect.Detector
	if options.ResolveLicenses {
		licenseDetector = local.NewDetector(logger)
	}

	generator, err := app.NewGenerator(options.ModuleDir,
		app.WithLogger(logger),
		app.WithIncludeFiles(options.IncludeFiles),
		app.WithIncludePackages(options.IncludePackages),
		app.WithIncludeStdlib(options.IncludeStd),
		app.WithLicenseDetector(licenseDetector),
		app.WithMainDir(options.Main))
	if err != nil {
		return err
	}

	bom, err := generator.Generate()
	if err != nil {
		return err
	}

	err = cliUtil.SetSerialNumber(bom, options.SBOMOptions)
	if err != nil {
		return fmt.Errorf("failed to set serial number: %w", err)
	}
	err = cliUtil.AddCommonMetadata(logger, bom)
	if err != nil {
		return fmt.Errorf("failed to add common metadata: %w", err)
	}
	if options.AssertLicenses {
		sbom.AssertLicenses(bom)
	}

	return cliUtil.WriteBOM(bom, options.OutputOptions)
}
