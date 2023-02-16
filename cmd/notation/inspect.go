package main

import (
	"crypto/sha1"
	b64 "encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"strings"
	"strconv"
	"time"

	"github.com/notaryproject/notation-core-go/signature"
	"github.com/notaryproject/notation-go/plugin/proto"
	"github.com/notaryproject/notation-go/registry"
	"github.com/notaryproject/notation/internal/cmd"
	"github.com/notaryproject/notation/internal/envelope"
	"github.com/notaryproject/notation/internal/ioutil"
	"github.com/notaryproject/notation/internal/tree"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/spf13/cobra"
)

type inspectOpts struct {
	cmd.LoggingFlagOpts
	SecureFlagOpts
	reference    string
	outputFormat string
}

type inspectOutput struct {
	MediaType  string `json:"mediaType"`
	Signatures []signatureOutput
}

type signatureOutput struct {
	MediaType             string              `json:"mediaType"`
	Digest                string              `json:"digest"`
	SignatureAlgorithm    string              `json:"signatureAlgorithm"`
	SignedAttributes      map[string]string   `json:"signedAttributes"`
	UserDefinedAttributes map[string]string   `json:"userDefinedAttributes"`
	UnsignedAttributes    map[string]string   `json:"unsignedAttributes"`
	Certificates          []certificateOutput `json:"certificates"`
	SignedArtifact        ocispec.Descriptor  `json:"signedArtifact"`
}

type certificateOutput struct {
	SHA1Fingerprint string `json:"SHA1Fingerprint"`
	IssuedTo        string `json:"issuedTo"`
	IssuedBy        string `json:"issuedBy"`
	Expiry          string `json:"expiry"`
}

func inspectCommand(opts *inspectOpts) *cobra.Command {
	if opts == nil {
		opts = &inspectOpts{}
	}
	command := &cobra.Command{
		Use:   "inspect [reference]",
		Short: "Inspect all signatures associated with the signed artifact",
		Long: `Inspect all signatures associated with the signed artifact.

Example - Inspect signatures on an OCI artifact identified by a digest:
  notation inspect <registry>/<repository>@<digest>

Example - Inspect signatures on an OCI artifact identified by a tag  (Notation will resolve tag to digest):
  notation inspect <registry>/<repository>:<tag>

Example - Inspect signatures on an OCI artifact identified by a digest and output as json:
  notation inspect --output json <registry>/<repository>@<digest>
`,
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return errors.New("missing reference")
			}
			opts.reference = args[0]
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return runInspect(cmd, opts)
		},
	}

	opts.LoggingFlagOpts.ApplyFlags(command.Flags())
	opts.SecureFlagOpts.ApplyFlags(command.Flags())
	cmd.SetPflagOutput(command.Flags(), &opts.outputFormat, cmd.PflagOutputUsage)
	return command
}

func runInspect(command *cobra.Command, opts *inspectOpts) error {
	// set log level
	ctx := opts.LoggingFlagOpts.SetLoggerLevel(command.Context())

	if opts.outputFormat != cmd.OutputJSON && opts.outputFormat != cmd.OutputPlaintext {
		return fmt.Errorf("unrecognized output format %s", opts.outputFormat)
	}

	// initialize
	reference := opts.reference
	sigRepo, err := getSignatureRepository(ctx, &opts.SecureFlagOpts, reference)
	if err != nil {
		return err
	}

	manifestDesc, ref, err := getManifestDescriptor(ctx, &opts.SecureFlagOpts, reference, sigRepo)
	if err != nil {
		return err
	}

	// reference is a digest reference
	if err := ref.ValidateReferenceAsDigest(); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: Always inspect the artifact using digest(@sha256:...) rather than a tag(:%s) because resolved digest may not point to the same signed artifact, as tags are mutable.\n", ref.Reference)
		ref.Reference = manifestDesc.Digest.String()
	}

	output := inspectOutput{MediaType: manifestDesc.MediaType, Signatures: []signatureOutput{}}
	skippedSignatures := false
	err = sigRepo.ListSignatures(ctx, manifestDesc, func(signatureManifests []ocispec.Descriptor) error {
		for _, sigManifestDesc := range signatureManifests {
			sigBlob, sigDesc, err := sigRepo.FetchSignatureBlob(ctx, sigManifestDesc)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Warning: unable to fetch signature %s due to error: %v\n", sigManifestDesc.Digest.String(), err)
				skippedSignatures = true
				continue
			}

			sigEnvelope, err := signature.ParseEnvelope(sigDesc.MediaType, sigBlob)
			if err != nil {
				logSkippedSignature(sigManifestDesc, err)
				skippedSignatures = true
				continue
			}

			envelopeContent, err := sigEnvelope.Content()
			if err != nil {
				logSkippedSignature(sigManifestDesc, err)
				skippedSignatures = true
				continue
			}

			signedArtifactDesc, err := envelope.DescriptorFromSignaturePayload(&envelopeContent.Payload)
			if err != nil {
				logSkippedSignature(sigManifestDesc, err)
				skippedSignatures = true
				continue
			}

			signatureAlgorithm, err := proto.EncodeSigningAlgorithm(envelopeContent.SignerInfo.SignatureAlgorithm)
			if err != nil {
				logSkippedSignature(sigManifestDesc, err)
				skippedSignatures = true
				continue
			}

			sig := signatureOutput{
				MediaType:             sigDesc.MediaType,
				Digest:                sigManifestDesc.Digest.String(),
				SignatureAlgorithm:    string(signatureAlgorithm),
				SignedAttributes:      getSignedAttributes(opts.outputFormat, envelopeContent),
				UserDefinedAttributes: signedArtifactDesc.Annotations,
				UnsignedAttributes:    getUnsignedAttributes(envelopeContent),
				Certificates:          getCertificates(opts.outputFormat, envelopeContent),
				SignedArtifact:        *signedArtifactDesc,
			}

			// clearing annotations from the SignedArtifact field since they're already
			// displayed as UserDefinedAttributes
			sig.SignedArtifact.Annotations = nil

			output.Signatures = append(output.Signatures, sig)
		}
		return nil
	})

	if err != nil {
		return err
	}

	err = printOutput(opts.outputFormat, ref.String(), output)
	if err != nil {
		return err
	}

	if skippedSignatures {
		return errors.New("at least one signature was skipped and not displayed")
	}

	return nil
}

func logSkippedSignature(sigDesc ocispec.Descriptor, err error) {
	fmt.Fprintf(os.Stderr, "Warning: Skipping signature %s because of error: %v\n", sigDesc.Digest.String(), err)
}

func getSignedAttributes(outputFormat string, envContent *signature.EnvelopeContent) map[string]string {
	signedAttributes := map[string]string{
		"signingScheme": string(envContent.SignerInfo.SignedAttributes.SigningScheme),
		"signingTime":   formatTimestamp(outputFormat, envContent.SignerInfo.SignedAttributes.SigningTime),
		"expiry":        formatTimestamp(outputFormat, envContent.SignerInfo.SignedAttributes.Expiry),
	}

	for _, attribute := range envContent.SignerInfo.SignedAttributes.ExtendedAttributes {
		signedAttributes[fmt.Sprint(attribute.Key)] = fmt.Sprint(attribute.Value)
	}

	return signedAttributes
}

func getUnsignedAttributes(envContent *signature.EnvelopeContent) map[string]string {
	unsignedAttributes := map[string]string{}

	if envContent.SignerInfo.UnsignedAttributes.TimestampSignature != nil {
		unsignedAttributes["timestampSignature"] = b64.StdEncoding.EncodeToString(envContent.SignerInfo.UnsignedAttributes.TimestampSignature)
	}

	if envContent.SignerInfo.UnsignedAttributes.SigningAgent != "" {
		unsignedAttributes["signingAgent"] = envContent.SignerInfo.UnsignedAttributes.SigningAgent
	}

	return unsignedAttributes
}

func formatTimestamp(outputFormat string, t time.Time) string {
	switch outputFormat {
	case cmd.OutputJSON:
		return t.Format(time.RFC3339)
	default:
		return t.Format(time.ANSIC)
	}
}

func getCertificates(outputFormat string, envContent *signature.EnvelopeContent) []certificateOutput {
	certificates := []certificateOutput{}

	for _, cert := range envContent.SignerInfo.CertificateChain {
		h := sha1.Sum(cert.Raw)
		fingerprint := strings.ToLower(hex.EncodeToString(h[:]))

		certificate := certificateOutput{
			SHA1Fingerprint: fingerprint,
			IssuedTo:        cert.Subject.String(),
			IssuedBy:        cert.Issuer.String(),
			Expiry:          formatTimestamp(outputFormat, cert.NotAfter),
		}

		certificates = append(certificates, certificate)
	}

	return certificates
}

func printOutput(outputFormat string, ref string, output inspectOutput) error {
	if outputFormat == cmd.OutputJSON {
		return ioutil.PrintObjectAsJSON(output)
	}

	fmt.Println("Inspecting all signatures for signed artifact")
	root := tree.New(ref)
	cncfSigNode := root.Add(registry.ArtifactTypeNotation)

	for _, signature := range output.Signatures {
		sigNode := cncfSigNode.Add(signature.Digest)
		sigNode.AddPair("media type", signature.MediaType)
		sigNode.AddPair("signature algorithm", signature.SignatureAlgorithm)

		signedAttributesNode := sigNode.Add("signed attributes")
		addMapToTree(signedAttributesNode, signature.SignedAttributes)

		userDefinedAttributesNode := sigNode.Add("user defined attributes")
		addMapToTree(userDefinedAttributesNode, signature.UserDefinedAttributes)

		unsignedAttributesNode := sigNode.Add("unsigned attributes")
		addMapToTree(unsignedAttributesNode, signature.UnsignedAttributes)

		certListNode := sigNode.Add("certificates")
		for _, cert := range signature.Certificates {
			certNode := certListNode.AddPair("SHA1 fingerprint", cert.SHA1Fingerprint)
			certNode.AddPair("issued to", cert.IssuedTo)
			certNode.AddPair("issued by", cert.IssuedBy)
			certNode.AddPair("expiry", cert.Expiry)
		}

		artifactNode := sigNode.Add("signed artifact")
		artifactNode.AddPair("media type", signature.SignedArtifact.MediaType)
		artifactNode.AddPair("digest", signature.SignedArtifact.Digest.String())
		artifactNode.AddPair("size", strconv.FormatInt(signature.SignedArtifact.Size, 10))
	}

	root.Print()
	return nil
}

func addMapToTree(node *tree.Node, m map[string]string) {
	if len(m) > 0 {
		for k, v := range m {
			node.AddPair(k, v)
		}
	} else {
		node.Add("(empty)")
	}
}
