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

package ioutil

import (
	"fmt"
	"time"

	"github.com/notaryproject/tspclient-go"
)

// Timestamp is a wrapper around tspclient.Timestamp to format the timestamp
// for display and JSON serialization.
type Timestamp tspclient.Timestamp

// MarshalJSON returns the Timestamp formatted in RFC3339.
func (t Timestamp) MarshalJSON() ([]byte, error) {
	ts := tspclient.Timestamp(t)
	return []byte(fmt.Sprintf("\"%s\"", ts.Format(time.RFC3339))), nil
}

// String returns the Timestamp formatted in ASNIC.
func (t Timestamp) String() string {
	ts := tspclient.Timestamp(t)
	return ts.Format(time.ANSIC)
}

// Time is a wrapper around time.Time to format the time.Time for display and
// JSON serialization.
type Time time.Time

// MarshalJSON returns the Time formatted in RFC3339.
func (t Time) MarshalJSON() ([]byte, error) {
	ts := time.Time(t)
	return []byte(fmt.Sprintf("\"%s\"", ts.Format(time.RFC3339))), nil
}

// String returns the Time formatted in ASNIC.
func (t Time) String() string {
	ts := time.Time(t)
	return ts.Format(time.ANSIC)
}
