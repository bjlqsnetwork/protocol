/*

 */

package keys

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/btcsuite/btcutil/base58"
	"github.com/pkg/errors"

	"github.com/Oneledger/protocol/utils"
)

type BTCMultiSig struct {
	Msg []byte `json:"msg"`

	M int `json:"m"`

	Signers []Address `json:"signers"`

	Signatures []BTCSignature `json:"signatures"`
}

type BTCSignature struct {
	Index   int     `json:"index"`
	Address Address `json:"address"` // this should be a compressed public key
	Sign    []byte  `json:"sign"`
}

func NewBTCMultiSig(msg []byte, m int, signers []Address) (*BTCMultiSig, error) {

	/*
		if msg == nil {
			return nil, ErrMissMsg
		}
	*/

	if signers == nil {
		return nil, ErrMissSigners
	}

	if m < 0 || m > len(signers) {
		return nil, ErrInvalidThreshold
	}

	return &BTCMultiSig{
		Msg:        msg,
		M:          m,
		Signers:    signers,
		Signatures: make([]BTCSignature, len(signers)),
	}, nil
}

func (m *BTCMultiSig) AddSignature(s *BTCSignature) error {

	if !bytes.Equal(s.Address, m.Signers[s.Index]) {
		return ErrNotExpectedSigner
	}

	// TODO verify the signature using btc libs

	m.Signatures[s.Index] = *s
	return nil
}

func (m BTCMultiSig) IsValid() bool {
	cnt := 0
	for _, item := range m.Signatures {
		if item.Sign != nil {
			cnt++
		}
	}
	if cnt < m.M {
		return false
	}
	return true
}

func (m BTCMultiSig) IsCancel() bool {
	cnt := 0
	for _, item := range m.Signatures {
		if item.Sign != nil {
			cnt++
		}
	}
	if cnt < m.M {
		return false
	}
	return true
}

func (m BTCMultiSig) Address() Address {
	s := &BTCMultiSig{m.Msg, m.M, m.Signers, nil}
	b := s.Bytes()
	return utils.Hash(b)
}

func (m BTCMultiSig) Bytes() []byte {

	signatures := m.Signatures
	m.Signatures = make([]BTCSignature, 0)

	for _, item := range signatures {
		if item.Sign != nil {
			m.Signatures = append(m.Signatures, item)
		}
	}

	b, _ := json.Marshal(m)
	return b
}

func (m BTCMultiSig) HasAddressSigned(addr Address) bool {
	index := len(m.Signers) + 100

	for i := range m.Signers {
		if bytes.Equal(addr, m.Signers[i]) {
			index = i
			break
		}
	}

	if index > len(m.Signers) {
		return false
	}

	if len(m.Signatures[index].Sign) == 0 {
		return false
	}

	return true
}

func (m *BTCMultiSig) FromBytes(b []byte) error {
	err := json.Unmarshal(b, m)
	if err != nil {
		return err
	}

	signatures := m.Signatures
	m.Signatures = make([]BTCSignature, len(m.Signers))
	for _, item := range signatures {
		m.Signatures[item.Index] = item
	}

	return nil
}

func (m *BTCMultiSig) GetSignerIndex(addr Address) (int, error) {

	for i := range m.Signers {
		if bytes.Equal(addr, m.Signers[i]) {
			return i, nil
		}
	}

	return 0, errors.New("address not found")
}

func (m *BTCMultiSig) GetSignatures() []BTCSignature {
	return m.Signatures
}

func (m *BTCMultiSig) GetSignaturesInOrder() [][]byte {

	tempSignatures := make([][]byte, len(m.Signers))
	for _, s := range m.Signatures {
		if len(s.Sign) == 0 {
			continue
		}

		index := s.Index
		tempSignatures[index] = s.Sign
		fmt.Println(s.Index, base58.Encode(s.Address))
	}

	signatures := make([][]byte, 0, len(m.Signers))
	for _, sig := range tempSignatures {
		if len(sig) == 0 {
			continue
		}

		signatures = append(signatures, sig)
	}

	return signatures
}

func (m *BTCMultiSig) Marshal() ([]byte, error) {
	return m.Bytes(), nil
}

func (m *BTCMultiSig) Unmarshal(b []byte) error {
	return m.FromBytes(b)
}
