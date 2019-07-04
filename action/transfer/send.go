package transfer

import (
	"fmt"

	"github.com/Oneledger/protocol/action"

	"github.com/Oneledger/protocol/data/balance"
	"github.com/pkg/errors"
	"github.com/tendermint/tendermint/libs/common"
)

var _ action.Msg = Send{}

type Send struct {
	From   action.Address
	To     action.Address
	Amount action.Amount
}

func (s Send) Signers() []action.Address {
	return []action.Address{s.From.Bytes()}
}

func (s Send) Type() action.Type {
	return action.SEND
}

var _ action.Tx = sendTx{}

type sendTx struct {
}

func (sendTx) Validate(ctx *action.Context, msg action.Msg, fee action.Fee, memo string, signatures []action.Signature) (bool, error) {
	//validate basic signature
	ok, err := action.ValidateBasic(msg, fee, memo, signatures)
	if err != nil {
		return ok, err
	}

	//validate transaction specific field
	send, ok := msg.(*Send)
	if !ok {
		return false, action.ErrWrongTxType
	}
	if !send.Amount.IsValid(ctx) {
		return false, errors.Wrap(action.ErrInvalidAmount, send.Amount.String())
	}
	if send.From == nil || send.To == nil {
		return false, action.ErrMissingData
	}
	return true, nil
}

func (sendTx) ProcessCheck(ctx *action.Context, msg action.Msg, fee action.Fee) (bool, action.Response) {
	ctx.Logger.Debug("Processing Send Transaction for CheckTx", msg, fee)
	balances := ctx.Balances

	send, ok := msg.(*Send)
	if !ok {
		return false, action.Response{Log: "failed to cast msg"}
	}

	b, _ := balances.Get(send.From.Bytes(), true)
	if b == nil {
		return false, action.Response{Log: "failed to get balance for sender"}
	}
	if !send.Amount.IsValid(ctx) {
		log := fmt.Sprint("amount is invalid", send.Amount, ctx.Currencies)
		return false, action.Response{Log: log}
	}
	coin := send.Amount.ToCoin(ctx)
	//check owner balance
	_, err := b.MinusCoin(coin)
	if err != nil {
		return false, action.Response{Log: err.Error()}
	}

	return true, action.Response{Tags: send.Tags()}
}

func (sendTx) ProcessDeliver(ctx *action.Context, msg action.Msg, fee action.Fee) (bool, action.Response) {
	ctx.Logger.Debug("Processing Send Transaction for DeliverTx", msg, fee)

	balances := ctx.Balances

	send, ok := msg.(*Send)
	if !ok {
		return false, action.Response{}
	}

	from, err := balances.Get(send.From.Bytes(), false)
	if err != nil {
		log := fmt.Sprint("Failed to get the balance of the owner ", send.From, "err", err)
		return false, action.Response{Log: log}
	}

	if !send.Amount.IsValid(ctx) {
		log := fmt.Sprint("amount is invalid", send.Amount, ctx.Currencies)
		return false, action.Response{Log: log}
	}

	coin := send.Amount.ToCoin(ctx)

	//change owner balance
	from, err = from.MinusCoin(coin)
	if err != nil {
		return false, action.Response{Log: err.Error()}
	}

	err = balances.Set(send.From.Bytes(), *from)
	if err != nil {
		log := fmt.Sprint("error updating balance in send transaction", err)
		return false, action.Response{Log: log}
	}

	//change receiver balance
	to, err := balances.Get(send.To.Bytes(), false)
	if err != nil {
		ctx.Logger.Error("failed to get the balance of the receipient", err)
	}
	if to == nil {
		to = balance.NewBalance()
	}
	to.AddCoin(coin)
	err = balances.Set(send.To.Bytes(), *to)
	if err != nil {
		return false, action.Response{Log: "balance set failed"}
	}
	return true, action.Response{Tags: send.Tags()}
}

func (sendTx) ProcessFee(ctx *action.Context, fee action.Fee) (bool, action.Response) {
	panic("implement me")
	// TODO: implement the fee charge for send
	return true, action.Response{GasWanted: 0, GasUsed: 0}
}

func (s Send) Tags() common.KVPairs {
	tags := make([]common.KVPair, 0)

	tag := common.KVPair{
		Key:   []byte("tx.type"),
		Value: []byte(s.Type().String()),
	}
	tag2 := common.KVPair{
		Key:   []byte("tx.owner"),
		Value: s.From.Bytes(),
	}
	tag3 := common.KVPair{
		Key:   []byte("tx.to"),
		Value: s.To.Bytes(),
	}

	tags = append(tags, tag, tag2, tag3)
	return tags
}