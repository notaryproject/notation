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

package scenario_test

import (
	"context"
	"net/http"
	"os"

	crlcore "github.com/notaryproject/notation-core-go/revocation/crl"
	"github.com/notaryproject/notation-go/verifier/crl"
	. "github.com/notaryproject/notation/test/e2e/internal/notation"
	"github.com/notaryproject/notation/test/e2e/internal/utils"
	. "github.com/notaryproject/notation/test/e2e/suite/common"
	. "github.com/onsi/ginkgo/v2"
)

var _ = Describe("notation CRL revocation check", Serial, func() {
	It("successfully completed with cache", func() {
		Host(CRLOptions(), func(notation *utils.ExecOpts, artifact *Artifact, vhost *utils.VirtualHost) {
			notation.Exec("sign", artifact.ReferenceWithDigest()).
				MatchKeyWords(SignSuccessfully)

			utils.LeafCRLUnrevoke()
			utils.IntermediateCRLUnrevoke()

			// verify without cache
			notation.Exec("verify", artifact.ReferenceWithDigest(), "-d").
				MatchKeyWords(
					VerifySuccessfully,
				).
				MatchErrKeyWords(
					"CRL file cache miss",
					"Retrieving crl bundle from file cache with key",
					"Storing crl bundle to file cache with key",
					"OCSP check failed with unknown error and fallback to CRL check for certificate #2",
				).
				NoMatchErrKeyWords(
					"is revoked",
				)

			// verify with cache
			notation.Exec("verify", artifact.ReferenceWithDigest(), "-d").
				MatchKeyWords(
					VerifySuccessfully,
				).
				MatchErrKeyWords(
					"Retrieving crl bundle from file cache with key",
					"OCSP check failed with unknown error and fallback to CRL check for certificate #2",
				).
				NoMatchErrKeyWords(
					"CRL file cache miss",
					"Storing crl bundle to file cache with key",
					"is revoked",
				)
		})
	})

	It("failed with revoked leaf certificate", func() {
		Host(CRLOptions(), func(notation *utils.ExecOpts, artifact *Artifact, vhost *utils.VirtualHost) {
			notation.Exec("sign", artifact.ReferenceWithDigest()).
				MatchKeyWords(SignSuccessfully)

			utils.LeafCRLRevoke()
			utils.IntermediateCRLUnrevoke()

			// verify without cache
			notation.ExpectFailure().Exec("verify", artifact.ReferenceWithDigest(), "-d").
				MatchErrKeyWords(
					VerifyFailed,
					"CRL file cache miss",
					"Retrieving crl bundle from file cache with key",
					"Storing crl bundle to file cache with key",
					"is revoked",
					"OCSP check failed with unknown error and fallback to CRL check for certificate #2",
				)

			// verify with cache
			notation.ExpectFailure().Exec("verify", artifact.ReferenceWithDigest(), "-d").
				MatchErrKeyWords(
					VerifyFailed,
					"Retrieving crl bundle from file cache with key",
					"is revoked",
					"OCSP check failed with unknown error and fallback to CRL check for certificate #2",
				).
				NoMatchErrKeyWords(
					"CRL file cache miss",
					"Storing crl bundle to file cache with key",
				)
		})
	})

	It("failed with revoked intermediate certificate", func() {
		Host(CRLOptions(), func(notation *utils.ExecOpts, artifact *Artifact, vhost *utils.VirtualHost) {
			notation.Exec("sign", artifact.ReferenceWithDigest()).
				MatchKeyWords(SignSuccessfully)

			utils.LeafCRLUnrevoke()
			utils.IntermediateCRLRevoke()

			// verify without cache
			notation.ExpectFailure().Exec("verify", artifact.ReferenceWithDigest(), "-d").
				MatchErrKeyWords(
					VerifyFailed,
					"CRL file cache miss",
					"Retrieving crl bundle from file cache with key",
					"Storing crl bundle to file cache with key",
					"is revoked",
					"OCSP check failed with unknown error and fallback to CRL check for certificate #2",
				)

			// verify with cache
			notation.ExpectFailure().Exec("verify", artifact.ReferenceWithDigest(), "-d").
				MatchErrKeyWords(
					VerifyFailed,
					"Retrieving crl bundle from file cache with key",
					"is revoked",
					"OCSP check failed with unknown error and fallback to CRL check for certificate #2",
				).
				NoMatchErrKeyWords(
					"CRL file cache miss",
					"Storing crl bundle to file cache with key",
				)
		})
	})

	It("successfully completed with cache creation error in warning message", func() {
		Host(CRLOptions(), func(notation *utils.ExecOpts, artifact *Artifact, vhost *utils.VirtualHost) {
			notation.Exec("sign", artifact.ReferenceWithDigest()).
				MatchKeyWords(SignSuccessfully)

			utils.LeafCRLUnrevoke()
			utils.IntermediateCRLUnrevoke()

			if err := os.MkdirAll(vhost.AbsolutePath(".cache"), 0500); err != nil {
				Fail(err.Error())
			}
			defer os.Chmod(vhost.AbsolutePath(".cache"), 0700)

			// verify without cache
			notation.Exec("verify", artifact.ReferenceWithDigest(), "-d").
				MatchKeyWords(
					VerifySuccessfully,
				).
				MatchErrKeyWords(
					"Warning: failed to create crl file cache",
					"OCSP check failed with unknown error and fallback to CRL check for certificate #2",
					`"GET" "http://localhost:10086/intermediate.crl"`,
					`"GET" "http://localhost:10086/leaf.crl"`,
				).
				NoMatchErrKeyWords(
					"is revoked",
				)
		})
	})

	It("failed with revoked leaf certificate and cache creation error in warning message", func() {
		Host(CRLOptions(), func(notation *utils.ExecOpts, artifact *Artifact, vhost *utils.VirtualHost) {
			notation.Exec("sign", artifact.ReferenceWithDigest()).
				MatchKeyWords(SignSuccessfully)

			utils.LeafCRLRevoke()
			utils.IntermediateCRLUnrevoke()

			if err := os.MkdirAll(vhost.AbsolutePath(".cache"), 0500); err != nil {
				Fail(err.Error())
			}
			defer os.Chmod(vhost.AbsolutePath(".cache"), 0700)

			// verify without cache
			notation.ExpectFailure().Exec("verify", artifact.ReferenceWithDigest(), "-d").
				MatchErrKeyWords(
					VerifyFailed,
					"Warning: failed to create crl file cache",
					"OCSP check failed with unknown error and fallback to CRL check for certificate #2",
					`"GET" "http://localhost:10086/intermediate.crl"`,
					`"GET" "http://localhost:10086/leaf.crl"`,
					"is revoked",
				)
		})
	})

	It("successfully completed with cache get and set error in debug log", func() {
		Host(CRLOptions(), func(notation *utils.ExecOpts, artifact *Artifact, vhost *utils.VirtualHost) {
			notation.Exec("sign", artifact.ReferenceWithDigest()).
				MatchKeyWords(SignSuccessfully)

			utils.LeafCRLUnrevoke()
			utils.IntermediateCRLUnrevoke()

			// verify without cache
			notation.Exec("verify", artifact.ReferenceWithDigest(), "-d").
				MatchKeyWords(
					VerifySuccessfully,
				).
				MatchErrKeyWords(
					"CRL file cache miss",
					"Retrieving crl bundle from file cache with key",
					"Storing crl bundle to file cache with key",
					"OCSP check failed with unknown error and fallback to CRL check for certificate #2",
				).
				NoMatchErrKeyWords(
					"is revoked",
				)

			utils.LeafCRLRevoke()
			if err := os.Chmod(vhost.AbsolutePath(".cache", "crl"), 0000); err != nil {
				Fail(err.Error())
			}
			defer os.Chmod(vhost.AbsolutePath(".cache", "crl"), 0700)

			// verify with cache error
			notation.ExpectFailure().Exec("verify", artifact.ReferenceWithDigest(), "-d").
				MatchErrKeyWords(
					VerifyFailed,
					"failed to get crl bundle from file cache with key",
					"failed to store crl bundle in file cache",
					"/.cache/crl/eaf8bbfe35f6c2c8b136081de9a994f9515752b2e30b9a6889ae3128ea97656c: permission denied",
					"OCSP check failed with unknown error and fallback to CRL check for certificate #2",
					`"GET" "http://localhost:10086/intermediate.crl"`,
					`"GET" "http://localhost:10086/leaf.crl"`,
					"is revoked",
				)
		})
	})

	It("succesfully completed with the crl in the cache expired and a roundtrip", func() {
		Host(CRLOptions(), func(notation *utils.ExecOpts, artifact *Artifact, vhost *utils.VirtualHost) {
			ctx := context.Background()
			notation.Exec("sign", artifact.ReferenceWithDigest()).
				MatchKeyWords(SignSuccessfully)

			utils.LeafCRLExpired()
			// write expired CRL cache
			fetcher, err := crlcore.NewHTTPFetcher(http.DefaultClient)
			if err != nil {
				Fail(err.Error())
			}
			fetcher.Cache, err = crl.NewFileCache(vhost.AbsolutePath(".cache", "crl"))
			if err != nil {
				Fail(err.Error())
			}
			_, err = fetcher.Fetch(ctx, "http://localhost:10086/leaf.crl")
			if err != nil {
				Fail(err.Error())
			}

			utils.LeafCRLUnrevoke()
			utils.IntermediateCRLUnrevoke()
			// verify without cache
			notation.Exec("verify", artifact.ReferenceWithDigest(), "-d").
				MatchKeyWords(
					VerifySuccessfully,
				).
				MatchErrKeyWords(
					"CRL bundle retrieved from file cache has expired at 2023-12-25",
					`"GET" "http://localhost:10086/leaf.crl"`,
					"Retrieving crl bundle from file cache with key",
					"Storing crl bundle to file cache with key",
					"OCSP check failed with unknown error and fallback to CRL check for certificate #2",
				).
				NoMatchErrKeyWords(
					"is revoked",
				)
		})
	})

	It("failed with crl in the cache expired and a roundtrip to download a revoked crl", func() {
		Host(CRLOptions(), func(notation *utils.ExecOpts, artifact *Artifact, vhost *utils.VirtualHost) {
			ctx := context.Background()
			notation.Exec("sign", artifact.ReferenceWithDigest()).
				MatchKeyWords(SignSuccessfully)

			utils.LeafCRLExpired()
			// write expired CRL cache
			fetcher, err := crlcore.NewHTTPFetcher(http.DefaultClient)
			if err != nil {
				Fail(err.Error())
			}
			fetcher.Cache, err = crl.NewFileCache(vhost.AbsolutePath(".cache", "crl"))
			if err != nil {
				Fail(err.Error())
			}
			_, err = fetcher.Fetch(ctx, "http://localhost:10086/leaf.crl")
			if err != nil {
				Fail(err.Error())
			}

			utils.LeafCRLRevoke()
			utils.IntermediateCRLUnrevoke()
			// verify without cache
			notation.ExpectFailure().Exec("verify", artifact.ReferenceWithDigest(), "-d").
				MatchErrKeyWords(
					VerifyFailed,
					"CRL bundle retrieved from file cache has expired at 2023-12-25",
					`"GET" "http://localhost:10086/leaf.crl"`,
					"Retrieving crl bundle from file cache with key",
					"Storing crl bundle to file cache with key",
					"OCSP check failed with unknown error and fallback to CRL check for certificate #2",
					"is revoked",
				)
		})
	})
})
