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
	"testing"

	"github.com/notaryproject/notation-go/log"
	"github.com/sirupsen/logrus"
)

func TestWithLoggerLevel(t *testing.T) {
	t.Run("debug level log", func(t *testing.T) {
		ctx := WithLoggerLevel(context.Background(), logrus.DebugLevel)
		logger := log.GetLogger(ctx)
		if logrusLogger, ok := logger.(*logrus.Logger); ok {
			if logrusLogger.Level != logrus.DebugLevel {
				t.Errorf("log level want = %v, got %v", logrus.DebugLevel, logrusLogger.Level)
			}
		} else {
			t.Fatal("should log with logrus")
		}
	})

	t.Run("info level log", func(t *testing.T) {
		ctx := WithLoggerLevel(context.Background(), logrus.InfoLevel)
		logger := log.GetLogger(ctx)
		if logrusLogger, ok := logger.(*logrus.Logger); ok {
			if logrusLogger.Level != logrus.InfoLevel {
				t.Errorf("log level want = %v, got %v", logrus.InfoLevel, logrusLogger.Level)
			}
		} else {
			t.Fatal("should log with logrus")
		}
	})
}
