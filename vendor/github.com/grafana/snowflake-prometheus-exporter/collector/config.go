// Copyright  Grafana Labs
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package collector

import (
	"errors"
	"fmt"
	"net/url"
)

type Config struct {
	AccountName string
	Username    string
	Password    string
	Role        string
	Warehouse   string
}

var (
	errNoAccountName = errors.New("account_name must be specified")
	errNoUsername    = errors.New("username must be specified")
	errNoPassword    = errors.New("password must be specified")
	errNoRole        = errors.New("role must be specified")
	errNoWarehouse   = errors.New("warehouse must be specified")
)

func (c Config) Validate() error {
	if c.AccountName == "" {
		return errNoAccountName
	}

	if c.Username == "" {
		return errNoUsername
	}

	if c.Password == "" {
		return errNoPassword
	}

	if c.Role == "" {
		return errNoRole
	}

	if c.Warehouse == "" {
		return errNoWarehouse
	}

	return nil
}

// snowflakeConnectionString returns a connection string to connect to the SNOWFLAKE database using the
// options specified in the config.
// Assumes the config is valid according to Validate().
func (c Config) snowflakeConnectionString() string {
	accountNameEscaped := url.QueryEscape(c.AccountName)
	usernameEscaped := url.QueryEscape(c.Username)
	passwordEscaped := url.QueryEscape(c.Password)
	roleEscaped := url.QueryEscape(c.Role)
	warehouseEscaped := url.QueryEscape(c.Warehouse)

	return fmt.Sprintf("%s:%s@%s/SNOWFLAKE?role=%s&warehouse=%s", usernameEscaped, passwordEscaped, accountNameEscaped, roleEscaped, warehouseEscaped)
}
