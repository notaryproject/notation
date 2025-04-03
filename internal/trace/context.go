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

	"github.com/notaryproject/notation-go/log"
	"github.com/sirupsen/logrus"
)

// WithLoggerLevel returns a context with logrus log entry.
func WithLoggerLevel(ctx context.Context, level logrus.Level) context.Context {
	// set formatter
	var formatter logrus.TextFormatter
	if level == logrus.DebugLevel {
		formatter.FullTimestamp = true
		// Set timestamp format to include nanoseconds and timezone
		formatter.TimestampFormat = "2006-01-02 15:04:05.000000000Z"
	} else {
		formatter.DisableTimestamp = true
	}

	// create logger
	logger := logrus.New()
	formatter.DisableQuote = true
	logger.SetFormatter(&formatter)
	logger.SetLevel(level)

	// Add UTC hook to convert timestamps to UTC
	logger.AddHook(&UTCHook{})

	return log.WithLogger(ctx, logger)
}

// UTCHook is a hook for logrus that converts timestamps to UTC
type UTCHook struct{}

// Levels returns the levels this hook is enabled for
func (h *UTCHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

// Fire is called by logrus when a log entry is created.
// This implementation converts the timestamp to UTC.
func (h *UTCHook) Fire(entry *logrus.Entry) error {
	entry.Time = entry.Time.UTC()
	return nil
}
