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
	"fmt"
	"time"

	"github.com/blinklabs-io/cardano-node-api/internal/config"

	"github.com/blinklabs-io/gouroboros/ledger"
	"github.com/blinklabs-io/gouroboros/protocol/chainsync"
	"github.com/blinklabs-io/gouroboros/protocol/common"
	"github.com/blinklabs-io/snek/event"
	input_chainsync "github.com/blinklabs-io/snek/input/chainsync"
)

func buildChainSyncConfig(connCfg ConnectionConfig) chainsync.Config {
	cfg := config.GetConfig()
	return chainsync.NewConfig(
		chainsync.WithBlockTimeout(
			time.Duration(cfg.Node.QueryTimeout)*time.Second,
		),
		// We wrap the handler funcs to include our ConnectionConfig
		chainsync.WithRollBackwardFunc(
			func(connCfg ConnectionConfig) chainsync.RollBackwardFunc {
				return func(ctx chainsync.CallbackContext, point common.Point, tip chainsync.Tip) error {
					return chainSyncRollBackwardHandler(
						ctx, connCfg, point, tip,
					)
				}
			}(connCfg),
		),
		chainsync.WithRollForwardFunc(
			func(connCfg ConnectionConfig) chainsync.RollForwardFunc {
				return func(ctx chainsync.CallbackContext, blockType uint, blockData any, tip chainsync.Tip) error {
					return chainSyncRollForwardHandler(
						ctx, connCfg, blockType, blockData, tip,
					)
				}
			}(connCfg),
		),
	)
}

func chainSyncRollBackwardHandler(
	ctx chainsync.CallbackContext,
	connCfg ConnectionConfig,
	point common.Point,
	tip chainsync.Tip,
) error {
	if connCfg.ChainSyncEventChan != nil {
		evt := event.New(
			"chainsync.rollback",
			time.Now(),
			nil,
			input_chainsync.NewRollbackEvent(point),
		)
		connCfg.ChainSyncEventChan <- evt
	}
	return nil
}

func chainSyncRollForwardHandler(
	ctx chainsync.CallbackContext,
	connCfg ConnectionConfig,
	blockType uint,
	blockData interface{},
	tip chainsync.Tip,
) error {
	cfg := config.GetConfig()
	if connCfg.ChainSyncEventChan != nil {
		var block ledger.Block
		switch v := blockData.(type) {
		case ledger.Block:
			block = v
		/*
			case ledger.BlockHeader:
				blockSlot := v.SlotNumber()
				blockHash, _ := hex.DecodeString(v.Hash())
				oConn, err := GetConnection()
				if err != nil {
					return err
				}
				block, err = oConn.BlockFetch().Client.GetBlock(common.NewPoint(blockSlot, blockHash))
				if err != nil {
					return err
				}
		*/
		default:
			return fmt.Errorf("unknown block data")
		}
		evt := event.New(
			"chainsync.block",
			time.Now(),
			input_chainsync.NewBlockContext(block, cfg.Node.NetworkMagic),
			input_chainsync.NewBlockEvent(block, true),
		)
		connCfg.ChainSyncEventChan <- evt
	}
	return nil
}
