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

// Package output provides the output tools for writing information to the
// output.
package output

import (
	"encoding/json"
	"fmt"
	"io"
	"sync"
)

// Printer prints for status handlers.
type Printer struct {
	out  io.Writer
	err  io.Writer
	lock sync.Mutex
}

// NewPrinter creates a new Printer.
func NewPrinter(out io.Writer, err io.Writer) *Printer {
	return &Printer{out: out, err: err}
}

// Write implements the io.Writer interface.
func (p *Printer) Write(b []byte) (int, error) {
	p.lock.Lock()
	defer p.lock.Unlock()

	return p.out.Write(b)
}

// Println prints objects concurrent-safely with newline.
func (p *Printer) Println(a ...any) error {
	p.lock.Lock()
	defer p.lock.Unlock()

	_, err := fmt.Fprintln(p.out, a...)
	if err != nil {
		err = fmt.Errorf("display output error: %w", err)
		_, _ = fmt.Fprint(p.err, err)
		return err
	}
	return nil
}

// Printf prints objects concurrent-safely.
func (p *Printer) Printf(format string, a ...any) error {
	p.lock.Lock()
	defer p.lock.Unlock()

	_, err := fmt.Fprintf(p.out, format, a...)
	if err != nil {
		err = fmt.Errorf("display output error: %w", err)
		_, _ = fmt.Fprint(p.err, err)
		return err
	}
	return nil
}

// ErrorPrintf prints objects to error output concurrent-safely.
func (p *Printer) ErrorPrintf(format string, a ...any) error {
	p.lock.Lock()
	defer p.lock.Unlock()

	_, err := fmt.Fprintf(p.err, format, a...)
	return err
}

// PrintPrettyJSON prints object to out in JSON format.
func PrintPrettyJSON(out io.Writer, object any) error {
	encoder := json.NewEncoder(out)
	encoder.SetIndent("", "  ")
	return encoder.Encode(object)
}
