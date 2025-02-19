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

// copied and adopted from https://github.com/oras-project/oras with
// modification
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

package option

import (
	"fmt"
	"slices"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// FormatType is the type of output format.
type FormatType string

// format types
var (
	// FormatTypeJSON is the JSON format type.
	FormatTypeJSON FormatType = "json"
	// FormatTypeText is the text format type for human-readable output.
	FormatTypeText FormatType = "text"
	// FormatTypeTree is the tree format type for human-readable output.
	FormatTypeTree FormatType = "tree"
)

// Format contains input and parsed options for formatted output flags.
type Format struct {
	CurrentType  string
	allowedTypes []FormatType
}

// ApplyFlags sets up the flags for the format option.
//
// The defaultType is the default format type.
// The otherTypes are additional format types that are allowed.
func (f *Format) ApplyFlags(fs *pflag.FlagSet, defaultType FormatType, otherTypes ...FormatType) {
	f.CurrentType = string(defaultType)
	f.allowedTypes = append(otherTypes, defaultType)

	var quotedAllowedTypes []string
	for _, t := range f.allowedTypes {
		quotedAllowedTypes = append(quotedAllowedTypes, fmt.Sprintf("'%s'", t))
	}
	usage := fmt.Sprintf("output format, options: %s", strings.Join(quotedAllowedTypes, ", "))
	// apply flags
	fs.StringVarP(&f.CurrentType, "output", "o", f.CurrentType, usage)
}

// Parse parses the input format flag.
func (opts *Format) Parse(_ *cobra.Command) error {
	if ok := slices.Contains(opts.allowedTypes, FormatType(opts.CurrentType)); !ok {
		return fmt.Errorf("invalid format type: %q", opts.CurrentType)
	}
	return nil
}
