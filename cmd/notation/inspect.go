package main

import (
	"crypto/sha256"
	b64 "encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/notaryproject/notation-core-go/signature"
	"github.com/notaryproject/notation-go/plugin/proto"
	"github.com/notaryproject/notation-go/registry"
	cmderr "github.com/notaryproject/notation/cmd/notation/internal/errors"
	"github.com/notaryproject/notation/cmd/notation/internal/experimental"
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
	reference         string
	outputFormat      string
	allowReferrersAPI bool
	maxSignatures     int
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
	SHA256Fingerprint string `json:"SHA256Fingerprint"`
	IssuedTo          string `json:"issuedTo"`
	IssuedBy          string `json:"issuedBy"`
	Expiry            string `json:"expiry"`
}

func inspectCommand(opts *inspectOpts) *cobra.Command {
	if opts == nil {
		opts = &inspectOpts{}
	}
	longMessage := `Inspect all signatures associated with the signed artifact.

Example - Inspect signatures on an OCI artifact identified by a digest:
  notation inspect <registry>/<repository>@<digest>

Example - Inspect signatures on an OCI artifact identified by a tag  (Notation will resolve tag to digest):
  notation inspect <registry>/<repository>:<tag>

Example - Inspect signatures on an OCI artifact identified by a digest and output as json:
  notation inspect --output json <registry>/<repository>@<digest>
`
	experimentalExamples := `
Example - [Experimental] Inspect signatures on an OCI artifact identified by a digest using the Referrers API, if not supported (returns 404), fallback to the Referrers tag schema
  notation inspect --allow-referrers-api <registry>/<repository>@<digest>
`
	command := &cobra.Command{
		Use:   "inspect [reference]",
		Short: "Inspect all signatures associated with the signed artifact",
		Long:  longMessage,
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return errors.New("missing reference")
			}
			opts.reference = args[0]
			return nil
		},
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return experimental.CheckFlagsAndWarn(cmd, "allow-referrers-api")
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if opts.maxSignatures <= 0 {
				return fmt.Errorf("max-signatures value %d must be a positive number", opts.maxSignatures)
			}
			return runInspect(cmd, opts)
		},
	}

	opts.LoggingFlagOpts.ApplyFlags(command.Flags())
	opts.SecureFlagOpts.ApplyFlags(command.Flags())
	cmd.SetPflagOutput(command.Flags(), &opts.outputFormat, cmd.PflagOutputUsage)
	command.Flags().IntVar(&opts.maxSignatures, "max-signatures", 100, "maximum number of signatures to evaluate or examine")
	cmd.SetPflagReferrersAPI(command.Flags(), &opts.allowReferrersAPI, fmt.Sprintf(cmd.PflagReferrersUsageFormat, "inspect"))
	experimental.HideFlags(command, experimentalExamples, []string{"allow-referrers-api"})
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
	sigRepo, err := getRemoteRepository(ctx, &opts.SecureFlagOpts, reference, opts.allowReferrersAPI)
	if err != nil {
		return err
	}
	manifestDesc, resolvedRef, err := resolveReferenceWithWarning(ctx, inputTypeRegistry, reference, sigRepo, "inspect")
	if err != nil {
		return err
	}
	output := inspectOutput{MediaType: manifestDesc.MediaType, Signatures: []signatureOutput{}}
	skippedSignatures := false
	err = listSignatures(ctx, sigRepo, manifestDesc, opts.maxSignatures, func(sigManifestDesc ocispec.Descriptor) error {
		sigBlob, sigDesc, err := sigRepo.FetchSignatureBlob(ctx, sigManifestDesc)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: unable to fetch signature %s due to error: %v\n", sigManifestDesc.Digest.String(), err)
			skippedSignatures = true
			return nil
		}

		sigEnvelope, err := signature.ParseEnvelope(sigDesc.MediaType, sigBlob)
		if err != nil {
			logSkippedSignature(sigManifestDesc, err)
			skippedSignatures = true
			return nil
		}

		envelopeContent, err := sigEnvelope.Content()
		if err != nil {
			logSkippedSignature(sigManifestDesc, err)
			skippedSignatures = true
			return nil
		}

		signedArtifactDesc, err := envelope.DescriptorFromSignaturePayload(&envelopeContent.Payload)
		if err != nil {
			logSkippedSignature(sigManifestDesc, err)
			skippedSignatures = true
			return nil
		}

		signatureAlgorithm, err := proto.EncodeSigningAlgorithm(envelopeContent.SignerInfo.SignatureAlgorithm)
		if err != nil {
			logSkippedSignature(sigManifestDesc, err)
			skippedSignatures = true
			return nil
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

		return nil
	})
	var errorExceedMaxSignatures cmderr.ErrorExceedMaxSignatures
	if err != nil && !errors.As(err, &errorExceedMaxSignatures) {
		return err
	}

	if err := printOutput(opts.outputFormat, resolvedRef, output); err != nil {
		return err
	}

	if errorExceedMaxSignatures.MaxSignatures > 0 {
		fmt.Println("Warning:", errorExceedMaxSignatures)
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
	}
	expiry := envContent.SignerInfo.SignedAttributes.Expiry
	if !expiry.IsZero() {
		signedAttributes["expiry"] = formatTimestamp(outputFormat, expiry)
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
		h := sha256.Sum256(cert.Raw)
		fingerprint := strings.ToLower(hex.EncodeToString(h[:]))

		certificate := certificateOutput{
			SHA256Fingerprint: fingerprint,
			IssuedTo:          cert.Subject.String(),
			IssuedBy:          cert.Issuer.String(),
			Expiry:            formatTimestamp(outputFormat, cert.NotAfter),
		}

		certificates = append(certificates, certificate)
	}

	return certificates
}

func printOutput(outputFormat string, ref string, output inspectOutput) error {
	if outputFormat == cmd.OutputJSON {
		return ioutil.PrintObjectAsJSON(output)
	}

	if len(output.Signatures) == 0 {
		fmt.Printf("%s has no associated signature\n", ref)
		return nil
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
			certNode := certListNode.AddPair("SHA256 fingerprint", cert.SHA256Fingerprint)
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
