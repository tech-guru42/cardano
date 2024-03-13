// Copyright 2024 Blink Labs Software
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

package node

import (
	"time"

	"github.com/blinklabs-io/gouroboros/protocol/localtxmonitor"

	"github.com/blinklabs-io/cardano-node-api/internal/config"
)

func buildLocalTxMonitorConfig() localtxmonitor.Config {
	cfg := config.GetConfig()
	return localtxmonitor.NewConfig(
		localtxmonitor.WithAcquireTimeout(
			time.Duration(cfg.Node.Timeout) * time.Second,
		),
		localtxmonitor.WithQueryTimeout(
			time.Duration(cfg.Node.QueryTimeout) * time.Second,
		),
	)
}
