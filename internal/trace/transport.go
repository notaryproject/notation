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

// Copied and adapted from oras (https://github.com/oras-project/oras)
/*
Copyright The ORAS Authors.
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package trace

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/notaryproject/notation-go/log"
	"github.com/sirupsen/logrus"
)

// Transport is an http.RoundTripper that keeps track of the in-flight
// request and add hooks to report HTTP tracing events.
type Transport struct {
	http.RoundTripper
}

func NewTransport(base http.RoundTripper) *Transport {
	return &Transport{base}
}

// RoundTrip calls base roundtrip while keeping track of the current request.
func (t *Transport) RoundTrip(req *http.Request) (resp *http.Response, err error) {
	ctx := req.Context()
	e := log.GetLogger(ctx)

	// logs to be printed out
	logs := fmt.Sprintf("> Request: %q %q\n", req.Method, req.URL)
	logs = logs + "> Request headers:\n"
	logs = logs + logHeader(req.Header)

	resp, err = t.RoundTripper.RoundTrip(req)
	if err != nil {
		e.Debugf(logs)
		e.Errorf("Error in getting response: %w", err)
	} else if resp == nil {
		e.Debugf(logs)
		e.Errorf("No response obtained for request %s %q", req.Method, req.URL)
	} else {
		logs = logs + fmt.Sprintf("< Response status: %q\n", resp.Status)
		logs = logs + "< Response headers:\n"
		logs = logs + logHeader(resp.Header)
		e.Debugf(logs)
	}
	return resp, err
}

// logHeader returns string of provided header keys and values, with auth header
// scrubbed.
func logHeader(header http.Header) string {
	if len(header) > 0 {
		var logs string
		for k, v := range header {
			if strings.EqualFold(k, "Authorization") {
				v = []string{"*****"}
			}
			logs = logs + fmt.Sprintf("   %q: %q\n", k, strings.Join(v, ", "))
		}
		return logs
	} else {
		return "   Empty header"
	}
}

// SetHTTPDebugLog sets up http debug log with logrus.Logger
func SetHTTPDebugLog(ctx context.Context, client *http.Client) *http.Client {
	if logrusLog, ok := log.GetLogger(ctx).(*logrus.Logger); !ok || logrusLog.Level != logrus.DebugLevel {
		return client
	}
	if client == nil {
		client = &http.Client{}
	}
	if client.Transport == nil {
		client.Transport = http.DefaultTransport
	}
	client.Transport = NewTransport(client.Transport)
	return client
}
