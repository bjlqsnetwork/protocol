/*
	Copyright 2017-2018 OneLedger
*/
package vm

import (
	"net/rpc"
	"strings"
	"time"

	"github.com/Oneledger/protocol/node/global"
	"github.com/Oneledger/protocol/node/log"
	"github.com/Oneledger/protocol/node/olvm/interpreter/runner"
	"github.com/Oneledger/protocol/node/sdk"
)

// TODO: Hardcoded port, needs to come from config
//var DefaultClient = NewClient("tcp", "localhost:1980")
var defaultClient *OLVMClient

func NewClient(protocol string, address string) *OLVMClient {

	return &OLVMClient{
		Protocol:    protocol,
		ServicePath: address,
	}
}

// Initialize the vm/daemon/etc.
func InitializeClient() {
	protocol := global.Current.OLVMProtocol
	address := global.Current.OLVMAddress

	defaultClient = NewClient(protocol, address)
}

func AutoRun(request *runner.OLVMRequest) (result *runner.OLVMResult, err error) {

	log.Debug("Trying to Run")
	result, err = defaultClient.Run(request)

	// TODO: Should be based on error code, not text...
	if err != nil {
		if strings.HasSuffix(err.Error(), "connection refused") {
			//try to launch the service

			// TODO: Not started the first time?
			log.Debug("Relaunching OLVM")
			go RunService()

			for err != nil && strings.HasSuffix(err.Error(), "connection refused") {
				time.Sleep(time.Second)
				log.Debug("Trying to ReRun")
				result, err = defaultClient.Run(request)
			}
		} else {
			log.Error("Run Failed", "err", err)
		}
		return
	}
	return
}

// Run a smart contract
func (c OLVMClient) Run(request *runner.OLVMRequest) (*runner.OLVMResult, error) {

	/*
		request := &runner.OLVMRequest{
			From:       from,
			Address:    address,
			CallString: callString,
			Value:      value,
		}
	*/

	client, err := rpc.DialHTTP(c.Protocol, ":"+sdk.GetPort(c.ServicePath))
	if err != nil {
		return nil, err
	}

	var result runner.OLVMResult
	// TODO: Shouldn't pass by address for the result
	err = client.Call("Container.Exec", request, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

func Run(request *runner.OLVMRequest) (*runner.OLVMResult, error) {
	return defaultClient.Run(request)
}
