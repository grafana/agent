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

package config

import (
	"os"
	"path/filepath"
	"testing"

	"go.uber.org/zap"

	"go.opentelemetry.io/collector/config/configmodels"
)

// LoadConfigFile loads a config from file.
func LoadConfigFile(t *testing.T, fileName string, factories Factories) (*configmodels.Config, error) {
	// Open the file for reading.
	file, err := os.Open(filepath.Clean(fileName))
	if err != nil {
		t.Error(err)
		return nil, err
	}

	defer func() {
		if errClose := file.Close(); errClose != nil {
			t.Error(errClose)
		}
	}()

	// Read yaml config from file
	v := NewViper()
	v.SetConfigType("yaml")
	err = v.ReadConfig(file)
	if err != nil {
		t.Errorf("unable to read yaml, %v", err)
		return nil, err
	}

	// Load the config from viper using the given factories.
	cfg, err := Load(v, factories)
	if err != nil {
		return nil, err
	}
	err = ValidateConfig(cfg, zap.NewNop())
	if err != nil {
		return nil, err
	}
	return cfg, nil
}
