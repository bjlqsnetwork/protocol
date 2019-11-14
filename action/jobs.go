/*

 */

package action

import (
	"github.com/Oneledger/protocol/data/ethereum"
	"os"

	"github.com/Oneledger/protocol/data/bitcoin"
	"github.com/Oneledger/protocol/data/keys"
	"github.com/Oneledger/protocol/log"
	"github.com/btcsuite/btcd/chaincfg"
)

type JobsContext struct {
	Service *Service
	Logger  *log.Logger

	Trackers *bitcoin.TrackerStore

	BTCPrivKey       keys.PrivateKey
	Params           *chaincfg.Params
	ValidatorAddress Address

	BlockCypherToken string

	LockScripts *bitcoin.LockScriptStore

	BTCNodeAddress string
	BTCRPCPort     string
	BTCRPCUsername string
	BTCRPCPassword string

	EthereumTrackers *ethereum.TrackerStore
	ETHContractABI string
	ETHConnection  string
	ETHContractAddress string


	BTCChainnet string

	// eth driver config
}

func NewJobsContext(chainType string, svc *Service, trackers *bitcoin.TrackerStore, privKey *keys.PrivateKey,
	valAddress keys.Address, bcyToken string, lStore *bitcoin.LockScriptStore,
	btcAddress, btcRPCPort, BTCRPCUsername, BTCRPCPassword, btcChain string,ETHAbi string,ETHconn string,ETHContractaddress string,ethTracker *ethereum.TrackerStore,
) *JobsContext {

	var params *chaincfg.Params
	switch chainType {
	case "mainnet":
		params = &chaincfg.MainNetParams
	case "testnet3":
		params = &chaincfg.TestNet3Params
	case "regtest":
		params = &chaincfg.RegressionNetParams
	case "simnet":
		params = &chaincfg.SimNetParams
	default:
		params = &chaincfg.TestNet3Params
	}

	w := os.Stdout

	return &JobsContext{
		Service:          svc,
		Logger:           log.NewLoggerWithPrefix(w, "internal_jobs"),
		Trackers:         trackers,
		BTCPrivKey:       *privKey,
		Params:           params,
		ValidatorAddress: valAddress,
		BlockCypherToken: bcyToken,
		LockScripts:      lStore,
		BTCNodeAddress: btcAddress,
		BTCRPCPort:     btcRPCPort,
		BTCRPCUsername: BTCRPCUsername,
		BTCRPCPassword: BTCRPCPassword,
		ETHConnection:ETHconn,
		ETHContractAddress:ETHContractaddress,
		ETHContractABI:ETHAbi,
		EthereumTrackers:ethTracker,
		BTCChainnet: btcChain,

	}

}
