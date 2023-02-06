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
