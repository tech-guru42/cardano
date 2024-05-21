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

package utxorpc

import (
	"context"
	"encoding/hex"
	"fmt"
	"log"

	connect "connectrpc.com/connect"
	"github.com/blinklabs-io/gouroboros/ledger"
	submit "github.com/utxorpc/go-codegen/utxorpc/v1alpha/submit"
	"github.com/utxorpc/go-codegen/utxorpc/v1alpha/submit/submitconnect"

	"github.com/blinklabs-io/cardano-node-api/internal/node"
)

// submitServiceServer implements the SubmitService API
type submitServiceServer struct {
	submitconnect.UnimplementedSubmitServiceHandler
}

// SubmitTx
func (s *submitServiceServer) SubmitTx(
	ctx context.Context,
	req *connect.Request[submit.SubmitTxRequest],
) (*connect.Response[submit.SubmitTxResponse], error) {

	// txRawList
	txRawList := req.Msg.GetTx() // []*AnyChainTx
	log.Printf("Got a SubmitTx request with %d transactions", len(txRawList))
	resp := &submit.SubmitTxResponse{}

	// Connect to node
	oConn, err := node.GetConnection(nil)
	if err != nil {
		return nil, err
	}
	defer func() {
		// Close Ouroboros connection
		oConn.Close()
	}()

	// Loop through the transactions and submit each
	errorList := make([]error, len(txRawList))
	hasError := false
	for i, txi := range txRawList {
		txRawBytes := txi.GetRaw() // raw bytes
		txType, err := ledger.DetermineTransactionType(txRawBytes)
		placeholderRef := []byte{}
		if err != nil {
			resp.Ref = append(resp.Ref, placeholderRef)
			errorList[i] = err
			hasError = true
			continue
		}
		tx, err := ledger.NewTransactionFromCbor(txType, txRawBytes)
		if err != nil {
			resp.Ref = append(resp.Ref, placeholderRef)
			errorList[i] = err
			hasError = true
			continue
		}
		// Submit the transaction
		err = oConn.LocalTxSubmission().Client.SubmitTx(
			uint16(txType),
			txRawBytes,
		)
		if err != nil {
			resp.Ref = append(resp.Ref, placeholderRef)
			errorList[i] = fmt.Errorf("%s", err.Error())
			hasError = true
			continue
		}
		txHexBytes, err := hex.DecodeString(tx.Hash())
		if err != nil {
			resp.Ref = append(resp.Ref, placeholderRef)
			errorList[i] = err
			hasError = true
			continue
		}
		resp.Ref = append(resp.Ref, txHexBytes)
	}
	if hasError {
		return connect.NewResponse(resp), fmt.Errorf("%v", errorList)
	}
	return connect.NewResponse(resp), nil
}

// WaitForTx

// ReadMempool
func (s *submitServiceServer) ReadMempool(
	ctx context.Context,
	req *connect.Request[submit.ReadMempoolRequest],
) (*connect.Response[submit.ReadMempoolResponse], error) {

	// This is GetTxs until https://github.com/utxorpc/spec/pull/95
	txim := req.Msg.GetTxs() // []*TxInMempool
	log.Printf("Got a ReadMempool request with %d transactions", len(txim))
	resp := &submit.ReadMempoolResponse{}

	// Connect to node
	oConn, err := node.GetConnection(nil)
	if err != nil {
		return nil, err
	}
	defer func() {
		// Close Ouroboros connection
		oConn.Close()
	}()

	// Start LocalTxMonitor client
	oConn.LocalTxMonitor().Client.Start()

	// Collect TX hashes from the mempool
	mempool := []*submit.TxInMempool{}
	for {
		txRawBytes, err := oConn.LocalTxMonitor().Client.NextTx()
		if err != nil {
			log.Printf("ERROR: %s", err)
			return nil, err
		}
		// No transactions in mempool
		if txRawBytes == nil {
			break
		}
		var act submit.AnyChainTx
		var actr submit.AnyChainTx_Raw
		actr.Raw = txRawBytes
		act.Type = &actr
		record := &submit.TxInMempool{
			Tx:    &act,
			Stage: submit.Stage_STAGE_MEMPOOL,
		}
		mempool = append(mempool, record)
	}

	// Check each requested Tx against our mempool
	for _, txi := range txim {
		txi.Stage = submit.Stage_STAGE_UNSPECIFIED
		for _, tx := range mempool {
			if txi.Stage == submit.Stage_STAGE_MEMPOOL {
				break
			}
			if txi.Tx.String() == tx.Tx.String() {
				txi.Stage = submit.Stage_STAGE_MEMPOOL
			}
		}
		resp.Stage = append(resp.Stage, txi.Stage)
	}
	return connect.NewResponse(resp), nil
}

// WatchMempool
