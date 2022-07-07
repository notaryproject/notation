package main

import (
	"fmt"
	"testing"
)

func TestVerifyCommand(t *testing.T) {
	command := verifyCommand()
	err := command.ParseFlags([]string{
		"ref",
		"--username", "uuu",
		"--password", "ppp",
		"-c", "c0",
		"-c", "c1",
		"--cert-file", "c0",
		"--cert-file", "c1",
		"--signature", "s0",
		"-s", "s1"})
	if err != nil {
		t.Fatalf("Parse Flag failed: %v", err)
	}
	if ref := command.Flags().Arg(0); ref != "ref" {
		t.Fatalf("Expect reference: %v, got: %v", "ref", ref)
	}
	if name, _ := command.Flags().GetString(flagUsername.Name); name != "uuu" {
		t.Fatalf("Expect %v: %v, got: %v", flagUsername.Name, "uuu", name)
	}
	if plain, _ := command.Flags().GetBool(flagPlainHTTP.Name); plain {
		t.Fatalf("Expect %v: %v, got: %v", flagPlainHTTP.Name, false, plain)
	}
	if password, _ := command.Flags().GetString(flagPassword.Name); password != "ppp" {
		t.Fatalf("Expect %v: %v, got: %v", flagPassword.Name, "ppp", password)
	}
	if pull, _ := command.Flags().GetBool("pull"); !pull {
		t.Fatalf("Expect pull: %v, got: %v", true, pull)
	}
	if mediaType, _ := command.Flags().GetString(flagMediaType.Name); mediaType != defaultMediaType {
		t.Fatalf("Expect %v: %v, got: %v", flagMediaType.Name, defaultMediaType, mediaType)
	}

	signatures, _ := command.Flags().GetStringSlice("signature")
	if len(signatures) != 2 {
		t.Fatalf("Expect signarure number: %v, got: %v", 2, len(signatures))
	}
	for i, signature := range signatures {
		if expected := fmt.Sprintf("s%v", i); expected != signature {
			t.Fatalf("Expect signature: %v, got: %v", expected, signature)
		}
	}

	certs, _ := command.Flags().GetStringSlice("cert")
	if len(certs) != 2 {
		t.Fatalf("Expect cert number: %v, got: %v", 2, len(certs))
	}
	for i, cert := range certs {
		if expected := fmt.Sprintf("c%v", i); expected != cert {
			t.Fatalf("Expect cert: %v, got: %v", expected, cert)
		}
	}

	certFiles, _ := command.Flags().GetStringSlice("cert-file")
	if len(certFiles) != 2 {
		t.Fatalf("Expect cert-file number: %v, got: %v", 2, len(certFiles))
	}
	for i, certFile := range certFiles {
		if expected := fmt.Sprintf("c%v", i); expected != certFile {
			t.Fatalf("Expect cert-file: %v, got: %v", expected, certFiles)
		}
	}
}

func TestVerifyWithMoreFlagCommand(t *testing.T) {
	command := verifyCommand()
	err := command.ParseFlags([]string{
		"ref",
		"--plain-http",
		"--pull=false",
		"--media-type=mmm"})

	if err != nil {
		t.Fatalf("Parse Flag failed: %v", err)
	}
	if ref := command.Flags().Arg(0); ref != "ref" {
		t.Fatalf("Expect reference: %v, got: %v", "ref", ref)
	}
	if plain, _ := command.Flags().GetBool(flagPlainHTTP.Name); !plain {
		t.Fatalf("Expect %v: %v, got: %v", flagPlainHTTP.Name, true, plain)
	}
	if pull, _ := command.Flags().GetBool("pull"); pull {
		t.Fatalf("Expect pull: %v, got: %v", false, pull)
	}
	if mediaType, _ := command.Flags().GetString(flagMediaType.Name); mediaType != "mmm" {
		t.Fatalf("Expect %v: %v, got: %v", flagMediaType.Name, "mmm", mediaType)
	}
}
