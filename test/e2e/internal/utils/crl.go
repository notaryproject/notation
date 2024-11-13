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

package utils

import (
	"fmt"
	"net/http"
)

// LeafCRLRevoke sends http post request to http://localhost:10086/leaf/revoke
func LeafCRLRevoke() error {
	url := "http://localhost:10086/leaf/revoke"
	resp, err := http.Post(url, "application/json", nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	fmt.Printf("CRL of leaf certificate revoked with status code: %d\n", resp.StatusCode)
	return nil
}

// LeafCRLUnrevoke sends http post request to http://localhost:10086/leaf/unrevoke
func LeafCRLUnrevoke() error {
	url := "http://localhost:10086/leaf/unrevoke"
	resp, err := http.Post(url, "application/json", nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	fmt.Printf("CRL of leaf certificate unrevoked with status code: %d\n", resp.StatusCode)
	return nil
}

// LeafCRLExpired sends http post request to http://localhost:10086/leaf/expired
func LeafCRLExpired() error {
	url := "http://localhost:10086/leaf/expired"
	resp, err := http.Post(url, "application/json", nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	fmt.Printf("CRL of leaf certificate expired with status code: %d\n", resp.StatusCode)
	return nil
}

// IntermediateCRLRevoke sends http post request to http://localhost:10086/intermediate/revoke
func IntermediateCRLRevoke() error {
	url := "http://localhost:10086/intermediate/revoke"
	resp, err := http.Post(url, "application/json", nil)
	if err != nil {
		return nil
	}
	defer resp.Body.Close()
	fmt.Printf("CRL of intermediate certificate revoked with status code: %d\n", resp.StatusCode)
	return nil
}

// IntermediateCRLUnrevoke sends http post request to http://localhost:10086/intermediate/unrevoke
func IntermediateCRLUnrevoke() error {
	url := "http://localhost:10086/intermediate/unrevoke"
	resp, err := http.Post(url, "application/json", nil)
	if err != nil {
		return nil
	}
	defer resp.Body.Close()
	fmt.Printf("CRL of intermediate certificate unrevoked with status code: %d\n", resp.StatusCode)
	return nil
}
