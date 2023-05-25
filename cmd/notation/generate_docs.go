// Copyright The Notary Project Authors.
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

package main

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/notaryproject/notation/internal/cmd"
	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc" // Add this line to import cobra doc lib
)

type generateDocsOpts struct {
	cmd.LoggingFlagOpts
	SecureFlagOpts
	outputDir string
}

func generateDocsCommand(opts *generateDocsOpts) *cobra.Command {
	if opts == nil {
		opts = &generateDocsOpts{
			outputDir: ".",
		}
	}
	command := &cobra.Command{
		Use:   "generateDocs <outputDir>",
		Short: "Generate reference documentation for Notation CLI",
		Long:  "Generate reference documentation for Notation CLI",
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return errors.New("no output directory specified")
			}
			opts.outputDir = args[0]
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return runGenerateDocs(cmd, opts)
		},
	}

	return command
}

func runGenerateDocs(cmd *cobra.Command, opts *generateDocsOpts) error {

	const fmTemplate = `---
title: "%s"
---

`

	filePrepender := func(filename string) string {
		name := strings.TrimSuffix(filename, filepath.Ext(filename))
		name = strings.TrimPrefix(name, opts.outputDir)
		return fmt.Sprintf(fmTemplate, strings.Replace(name, "_", " ", -1))
	}

	linkHandler := func(filename string) string {
		link := strings.TrimSuffix(filename, filepath.Ext(filename))
		return fmt.Sprintf("{{< ref \"/docs/cli-reference/%s\" >}}", link)
	}

	fmt.Println("Generating reference documentation for Notation CLI...")
	if err := doc.GenMarkdownTreeCustom(cmd.Root(), opts.outputDir, filePrepender, linkHandler); err != nil {
		return err
	}
	fmt.Println("Reference documentation for Notation CLI generated successfully at", opts.outputDir)

	return nil
}
