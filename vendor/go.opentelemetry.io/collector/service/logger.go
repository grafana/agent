// Copyright The OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package service

import (
	"flag"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"go.opentelemetry.io/collector/internal/version"
)

const (
	logLevelCfg   = "log-level"
	logProfileCfg = "log-profile"
)

var (
	// Command line pointer to logger level flag configuration.
	loggerLevelPtr   *string
	loggerProfilePtr *string
)

func loggerFlags(flags *flag.FlagSet) {
	loggerLevelPtr = flags.String(logLevelCfg, "INFO", "Output level of logs (DEBUG, INFO, WARN, ERROR, DPANIC, PANIC, FATAL)")
	loggerProfilePtr = flags.String(logProfileCfg, "", "Logging profile to use (dev, prod)")
}

func newLogger(hooks ...func(zapcore.Entry) error) (*zap.Logger, error) {
	var level zapcore.Level
	err := (&level).UnmarshalText([]byte(*loggerLevelPtr))
	if err != nil {
		return nil, err
	}

	conf := zap.NewProductionConfig()

	// Use logger profile if set on command line before falling back
	// to default based on build type.
	switch *loggerProfilePtr {
	case "dev":
		conf = zap.NewDevelopmentConfig()
	case "prod":
		conf = zap.NewProductionConfig()
	default:
		if version.IsDevBuild() {
			conf = zap.NewDevelopmentConfig()
		}
	}

	conf.Level.SetLevel(level)
	return conf.Build(zap.Hooks(hooks...))
}
