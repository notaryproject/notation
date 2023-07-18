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
	"fmt"
	"reflect"
)

func fieldSet(name ...string) map[string]struct{} {
	fs := make(map[string]struct{})
	for _, n := range name {
		fs[n] = struct{}{}
	}
	return fs
}

func validateRequiredField(data any, fieldSet map[string]struct{}) error {
	valueOfData := reflect.ValueOf(data)
	typeOfData := valueOfData.Type()
	for i := 0; i < valueOfData.NumField(); i++ {
		fieldName := typeOfData.Field(i).Name
		if _, ok := fieldSet[fieldName]; ok {
			// check the required field
			if valueOfData.Field(i).IsZero() {
				return fmt.Errorf("%s is not set", fieldName)
			}
		}
	}
	return nil
}
