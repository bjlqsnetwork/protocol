package event

import (
	"github.com/Oneledger/protocol/serialize"
	"github.com/Oneledger/protocol/utils/transition"
)

var (
	EthEngine transition.Engine
	BtcEngine transition.Engine
)

const (
	JobTypeAddSignature     = "addSignature"
	JobTypeBTCBroadcast     = "btcBroadcast"
	JobTypeBTCCheckFinality = "btcCheckFinality"
	JobTypeETHCheckfinalty  = "ethCheckFinality"
	JobTypeETHBroadcast     = "ethBroadcast"

	MaxJobRetries = 10
)

func init() {
	serialize.RegisterConcrete(new(JobAddSignature), "btc_addsign")
	serialize.RegisterConcrete(new(JobBTCBroadcast), "btc_broadcast")
	serialize.RegisterConcrete(new(JobBTCCheckFinality), "btc_cf")
	serialize.RegisterConcrete(new(JobETHBroadcast), "eth_broadcast")
	serialize.RegisterConcrete(new(JobETHCheckFinality), "eth_cf")
}
