package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/notaryproject/nv2/cmd/docker-nv2/config"
	"github.com/urfave/cli/v2"
)

var certsCommand = &cli.Command{
	Name:  "certificates",
	Usage: "Manage certificates used for signing and verification",
	Subcommands: []*cli.Command{
		certsAddCommand,
		certsListCommand,
		certsRemoveCommand,
	},
}

var certsAddCommand = &cli.Command{
	Name:      "add",
	Usage:     "Add certificate to verification list",
	ArgsUsage: "[cert]",

	Action: addCert,
}

var certsListCommand = &cli.Command{
	Name:    "list",
	Usage:   "List certificates used for verification",
	Aliases: []string{"ls"},
	Action:  listCerts,
}

var certsRemoveCommand = &cli.Command{
	Name:      "remove",
	Usage:     "Remove certificate from verification list",
	Aliases:   []string{"rm"},
	ArgsUsage: "[cert]",
	Action:    removeCerts,
}

func uniqueAppend(entries []string, e string) []string {
	entries = append(entries, e)
	keys := make(map[string]bool)
	list := []string{}
	for _, item := range entries {
		if _, value := keys[item]; !value {
			keys[item] = true
			list = append(list, item)
		}
	}
	return list
}

func uniqueRemove(entries []string, e string) ([]string, error) {
	keys := make(map[string]bool)
	list := []string{}
	found := false
	for _, item := range entries {
		if item == e {
			keys[item] = true
			found = true
		} else if _, value := keys[item]; !value {
			keys[item] = true
			list = append(list, item)
		}
	}

	if !found {
		return nil, fmt.Errorf("%s not in the list", e)
	}

	return list, nil
}

func addCert(ctx *cli.Context) error {

	if !ctx.Args().Present() {
		return errors.New("Required argument, certificate path not specified")
	}
	cert := ctx.Args().First()

	cfg, err := config.Load()
	if err != nil {
		if !os.IsNotExist(err) {
			return err
		}
		cfg = config.New()
	}
	cfg.VerificationCerts = uniqueAppend(cfg.VerificationCerts, cert)

	err = cfg.Save()
	if err == nil {
		fmt.Printf("Added %s to verification certificates\n", cert)
	}

	return nil
}

func listCerts(ctx *cli.Context) error {

	cfg, err := config.Load()
	if err != nil {
		if !os.IsNotExist(err) {
			return nil
		}
	}

	for _, s := range cfg.VerificationCerts {
		fmt.Printf("%s\n", s)
	}
	return nil
}

func removeCerts(ctx *cli.Context) error {

	if !ctx.Args().Present() {
		return errors.New("Required argument, certificate path not specified")
	}
	cert := ctx.Args().First()

	cfg, err := config.Load()
	if err != nil {
		if !os.IsNotExist(err) {
			return err
		}
		cfg = config.New()
	}
	cfg.VerificationCerts, err = uniqueRemove(cfg.VerificationCerts, cert)
	if err != nil {
		return err
	}

	err = cfg.Save()
	if err == nil {
		fmt.Printf("Removed %s from list of verification certificates\n", cert)
	}

	return nil
}
