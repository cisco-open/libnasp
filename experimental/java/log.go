//  Copyright (c) 2023 Cisco and/or its affiliates. All rights reserved.
//
//   Licensed under the Apache License, Version 2.0 (the "License");
//   you may not use this file except in compliance with the License.
//   You may obtain a copy of the License at
//
//        https://www.apache.org/licenses/LICENSE-2.0
//
//   Unless required by applicable law or agreed to in writing, software
//   distributed under the License is distributed on an "AS IS" BASIS,
//   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//   See the License for the specific language governing permissions and
//   limitations under the License.
//

package nasp

import (
	"code.cloudfoundry.org/go-diodes"
	"github.com/go-logr/logr"
	"github.com/go-logr/zerologr"
	"github.com/rs/zerolog"
)

// LogLevel defines log levels.
type LogLevel int

const (
	// TraceLogLevel defines trace log level.
	TraceLogLevel LogLevel = iota

	// DebugLogLevel defines debug log level.
	DebugLogLevel

	// InfoLogLevel defines info log level.
	InfoLogLevel

	// WarnLogLevel defines warn log level.
	WarnLogLevel

	// ErrorLogLevel defines error log level.
	ErrorLogLevel
)

var logOutput diodes.Diode
var logger logr.Logger

func setupLogger(logLevel LogLevel) {
	logOutput = diodes.NewManyToOne(1000, nil)
	zlLogLevel := zerolog.InfoLevel

	switch logLevel {
	case TraceLogLevel:
		zlLogLevel = zerolog.TraceLevel
	case DebugLogLevel:
		zlLogLevel = zerolog.DebugLevel
	case InfoLogLevel:
		zlLogLevel = zerolog.InfoLevel
	case WarnLogLevel:
		zlLogLevel = zerolog.WarnLevel
	case ErrorLogLevel:
		zlLogLevel = zerolog.ErrorLevel
	}

	zl := zerolog.New(newLogOutputWriter(logOutput)).
		With().Caller().Logger().
		Level(zlLogLevel)

	logger = zerologr.New(&zl)
}

// NextLogBatchJSON return the next log lines batch in json format if available otherwise nil
func NextLogBatchJSON(batchSize int) []byte {
	if batchSize <= 0 {
		return nil
	}

	var logBatch []byte
	for i := 0; i < batchSize; i++ {
		logLine, ok := logOutput.TryNext()
		if !ok {
			break
		}

		if i > 0 {
			logBatch = append(logBatch, byte(','))
		}

		logBatch = append(logBatch, *(*[]byte)(logLine)...)
	}

	if len(logBatch) == 0 {
		return nil
	}

	logBatchJSON := make([]byte, 0, len(logBatch)+2)
	logBatchJSON = append(logBatchJSON, byte('['))
	logBatchJSON = append(logBatchJSON, logBatch...)
	logBatchJSON = append(logBatchJSON, byte(']'))

	return logBatchJSON
}

type logOutputWriter struct {
	logOutput diodes.Diode
}

func (o *logOutputWriter) Write(p []byte) (n int, err error) {
	logLine := make([]byte, len(p))
	copy(logLine, p)

	o.logOutput.Set(diodes.GenericDataType(&logLine))
	return len(p), nil
}

func newLogOutputWriter(logOutput diodes.Diode) *logOutputWriter {
	return &logOutputWriter{
		logOutput: logOutput,
	}
}
