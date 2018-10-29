/*
	Copyright 2017 - 2018 OneLedger

	Query the chain for answers
*/
package shared

import (
	"bytes"
	"encoding/hex"
	"github.com/Oneledger/protocol/node/convert"
	ctypes "github.com/tendermint/tendermint/rpc/core/types"

	"github.com/Oneledger/protocol/node/action"
	"github.com/Oneledger/protocol/node/comm"
	"github.com/Oneledger/protocol/node/data"
	"github.com/Oneledger/protocol/node/id"
	"github.com/Oneledger/protocol/node/log"
	"github.com/Oneledger/protocol/node/serial"
)

func GetAccountKey(identity string) []byte {
	request := action.Message("Identity=" + identity)
	response := comm.Query("/accountKey", request)

	if response == nil || response.Response.Value == nil {
		log.Error("No Response from Node", "identity", identity)
		return nil
	}

	value := response.Response.Value
	if value == nil || len(value) == 0 {
		log.Error("Key is Missing", "identity", identity)
		return nil
	}

	key, status := hex.DecodeString(string(value))
	if status != nil {
		log.Error("Decode Failed", "identity", identity, "value", value)
		return nil
	}

	return key
}

func GetSwapAddress(currencyName string) []byte {
	request := action.Message("currency=" + currencyName)
	response := comm.Query("/swapAddress", request)

	if response == nil || response.Response.Value == nil {
		log.Fatal("Failed to get address", "chain", currencyName)
	}

	value := response.Response.Value
	if value == nil || len(value) == 0 {
		log.Fatal("Returned address is empty", "chain", currencyName)
	}

	return value
}

func GetBalance(accountKey id.AccountKey) *data.Coin {
	request := action.Message("accountKey=" + hex.EncodeToString(accountKey))

	// Send out a query
	response := comm.Query("/balance", request)
	if response == nil {
		log.Error("Failed to get Balance", "response", response, "key", accountKey)
		return nil
	}

	// Check the response
	value := response.Response.Value
	if value == nil || len(value) == 0 {
		log.Error("Failed to return Balance", "response", response, "key", accountKey)
		return nil
	}
	if bytes.Compare(value, []byte("null")) == 0 {
		log.Error("Null Balance", "response", response, "key", "accountKey")
		return nil
	}

	// Convert to a balance
	var balance data.Balance
	buffer, status := serial.Deserialize(value, &balance, serial.CLIENT)
	if status != nil {
		log.Error("Deserialize", "status", status, "value", value)
		return nil
	}
	if buffer == nil {
		log.Error("Can't deserialize", "response", response)
		return nil
	}

	log.Debug("Deserialize", "buffer", buffer, "response", response, "value", value)
	return &(buffer.(*data.Balance).Amount)
}

func GetTxByHash(hash []byte) *ctypes.ResultTx {
	response := comm.Tx(hash, true)
	if response == nil {
		log.Error("Search tx by hash failed", "hash", hash)
		return nil
	}
	return response
}

func GetTxByHeight(height int) *ctypes.ResultTxSearch {
	request := "tx.height=" + convert.GetString(height)

	response := comm.Search(request, true, 1, 100)
	if response == nil {
		log.Error("Search tx by height failed", "request", request)
	}
	return response
}

func GetTxByType(t string) *ctypes.ResultTxSearch {
	request := "tx.type" + t

	response := comm.Search(request, true, 1, 100)
	if response == nil {
		log.Error("Search tx by hash failed", "request", request)
		return nil
	}
	return response
}
