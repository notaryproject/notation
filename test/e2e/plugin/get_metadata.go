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
	"github.com/notaryproject/notation-go/plugin/proto"
	"github.com/notaryproject/notation/test/e2e/plugin/internal/io"
	"github.com/spf13/cobra"
)

func getPluginMetadataCommand() *cobra.Command {
	return &cobra.Command{
		Use: "get-plugin-metadata",
		RunE: func(cmd *cobra.Command, args []string) error {
			req := &proto.GetMetadataRequest{}
			if err := io.UnmarshalRequest(req); err != nil {
				return err
			}
			return runGetPluginMetadata(req)
		},
	}
}

func runGetPluginMetadata(req *proto.GetMetadataRequest) error {
	resp := &proto.GetMetadataResponse{
		Name:                      "e2e-plugin",
		Description:               "The e2e-plugin is a Notation compatible plugin for Notation E2E test",
		Version:                   "1.0.0",
		URL:                       "https://github.com/notaryproject/notation/test/e2e/plugin",
		SupportedContractVersions: []string{"1.0"},
		Capabilities: []proto.Capability{
			proto.CapabilityTrustedIdentityVerifier,
			proto.CapabilityRevocationCheckVerifier,
		},
	}

	// enable signing capability by PluginConfig
	checkCapability(req.PluginConfig, proto.CapabilitySignatureGenerator, resp)
	checkCapability(req.PluginConfig, proto.CapabilityEnvelopeGenerator, resp)

	// output the response
	return io.PrintResponse(resp)
}

func checkCapability(pluginConfig map[string]string, capability proto.Capability, resp *proto.GetMetadataResponse) {
	if v, ok := pluginConfig[string(capability)]; ok && v == "true" {
		resp.Capabilities = append(resp.Capabilities, capability)
	}
}
