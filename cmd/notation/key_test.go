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
	"reflect"
	"testing"
)

func TestKeyAddCommand_BasicArgs(t *testing.T) {
	opts := &keyAddOpts{}
	cmd := keyAddCommand(opts)
	expected := &keyAddOpts{
		name:         "name",
		plugin:       "pluginname",
		id:           "pluginid",
		pluginConfig: []string{"pluginconfig"},
	}
	if err := cmd.ParseFlags([]string{
		"--plugin", expected.plugin,
		"--id", expected.id,
		"--plugin-config", "pluginconfig",
		expected.name}); err != nil {
		t.Fatalf("Parse Flag failed: %v", err)
	}
	if err := cmd.Args(cmd, cmd.Flags().Args()); err != nil {
		t.Fatalf("Parse Args failed: %v", err)
	}
	if !reflect.DeepEqual(*expected, *opts) {
		t.Fatalf("Expect key add opts: %v, got: %v", expected, opts)
	}
}

func TestKeyUpdateCommand_BasicArgs(t *testing.T) {
	opts := &keyUpdateOpts{}
	cmd := keyUpdateCommand(opts)
	expected := &keyUpdateOpts{
		name:      "name",
		isDefault: true,
	}
	if err := cmd.ParseFlags([]string{
		expected.name,
		"--default"}); err != nil {
		t.Fatalf("Parse Flag failed: %v", err)
	}
	if err := cmd.Args(cmd, cmd.Flags().Args()); err != nil {
		t.Fatalf("Parse Args failed: %v", err)
	}
	if *expected != *opts {
		t.Fatalf("Expect key update opts: %v, got: %v", expected, opts)
	}
}

func TestKeyUpdateCommand_MissingArgs(t *testing.T) {
	cmd := keyUpdateCommand(nil)
	if err := cmd.ParseFlags(nil); err != nil {
		t.Fatalf("Parse Flag failed: %v", err)
	}
	if err := cmd.Args(cmd, cmd.Flags().Args()); err == nil {
		t.Fatal("Parse Args expected error, but ok")
	}
}

func TestKeyRemoveCommand_BasicArgs(t *testing.T) {
	opts := &keyDeleteOpts{}
	cmd := keyDeleteCommand(opts)
	expected := &keyDeleteOpts{
		names: []string{"key0", "key1", "key2"},
	}
	if err := cmd.ParseFlags(expected.names); err != nil {
		t.Fatalf("Parse Flag failed: %v", err)
	}
	if err := cmd.Args(cmd, cmd.Flags().Args()); err != nil {
		t.Fatalf("Parse Args failed: %v", err)
	}
	if !reflect.DeepEqual(*expected, *opts) {
		t.Fatalf("Expect key remove opts: %v, got: %v", expected, opts)
	}
}

func TestKeyRemoveCommand_MissingArgs(t *testing.T) {
	cmd := keyDeleteCommand(nil)
	if err := cmd.ParseFlags(nil); err != nil {
		t.Fatalf("Parse Flag failed: %v", err)
	}
	if err := cmd.Args(cmd, cmd.Flags().Args()); err == nil {
		t.Fatal("Parse Args expected error, but ok")
	}
}
