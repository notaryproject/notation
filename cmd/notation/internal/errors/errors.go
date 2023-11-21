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

package errors

import "fmt"

// ErrorReferrersAPINotSupported is used when the target registry does not
// support the Referrers API
type ErrorReferrersAPINotSupported struct {
	Msg string
}

func (e ErrorReferrersAPINotSupported) Error() string {
	if e.Msg != "" {
		return e.Msg
	}
	return "referrers API not supported"
}

// ErrorOCILayoutMissingReference is used when signing local content in oci
// layout folder but missing input tag or digest.
type ErrorOCILayoutMissingReference struct {
	Msg string
}

func (e ErrorOCILayoutMissingReference) Error() string {
	if e.Msg != "" {
		return e.Msg
	}
	return "reference is missing either digest or tag"
}

// ErrorExceedMaxSignatures is used when the number of signatures has surpassed
// the maximum limit that can be evaluated.
type ErrorExceedMaxSignatures struct {
	MaxSignatures int
}

func (e ErrorExceedMaxSignatures) Error() string {
	return fmt.Sprintf("exceeded configured limit of max signatures %d to examine", e.MaxSignatures)
}
