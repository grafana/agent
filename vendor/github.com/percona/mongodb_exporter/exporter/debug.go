// mongodb_exporter
// Copyright (C) 2022 Percona LLC
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <https://www.gnu.org/licenses/>.

package exporter

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/sirupsen/logrus"
)

func debugResult(log *logrus.Logger, m interface{}) {
	if !log.IsLevelEnabled(logrus.DebugLevel) {
		return
	}

	debugStr, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		log.Errorf("cannot marshal struct for debug: %s", err)
		return
	}

	// don't use logrus because:
	// 1. It will escape new lines and " making it harder to read and to use
	// 2. It will add timestamp
	// 3. This way is easier to copy/paste to put the info in a ticket
	fmt.Fprintln(os.Stderr, string(debugStr))
}
