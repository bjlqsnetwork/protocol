/*
	Copyright 2017-2018 OneLedger

	OneLedger Wallet
*/
package olwallet


import (
	"crypto/sha256"
	"github.com/btcsuite/btcd/btcec"
	//"../channeldb"
)

// OneLedgerChannel implements the state machine which corresponds to the current commitment protocol wire spec.
type OneLedgerChannel struct {

}

// PaymentHash represents the sha256 of a random value. This hash is used to
// uniquely track incoming/outgoing payments within this channel, as well as
// payments requested by the wallet/daemon.
type PaymentHash [32]byte

// updateType is the exact type of an entry within the shared HTLC log.
type updateType uint8

// PaymentDescriptor represents a commitment state update which either adds,
// settles, or removes an HTLC. PaymentDescriptors encapsulate all necessary
// metadata w.r.t to an HTLC, and additional data pairing a settle message to
// the original added HTLC.
type PaymentDescriptor struct {
	// RHash is the payment hash for this HTLC. The HTLC can be settled iff
	// the preimage to this hash is presented.
	RHash PaymentHash

	// RPreimage is the preimage that settles the HTLC pointed to within the
	// log by the ParentIndex.
	RPreimage PaymentHash

	// Timeout is the absolute timeout in blocks, after which this HTLC
	// expires.
	Timeout uint32

	// Amount is the HTLC amount.
	Amount uint64

	// LogIndex is the log entry number that his HTLC update has within the
	// log. Depending on if IsIncoming is true, this is either an entry the
	// remote party added, or one that we added locally.
	LogIndex uint64

	// HtlcIndex is the index within the main update log for this HTLC.
	// Entries within the log of type Add will have this field populated,
	// as other entries will point to the entry via this counter.
	//
	// NOTE: This field will only be populate if EntryType is Add.
	HtlcIndex uint64

	// ParentIndex is the HTLC index of the entry that this update settles or
	// times out.
	//
	// NOTE: This field will only be populate if EntryType is Fail or
	// Settle.
	ParentIndex uint64

	// localOutputIndex is the output index of this HTLc output in the
	// commitment transaction of the local node.
	//
	// NOTE: If the output is dust from the PoV of the local commitment
	// chain, then this value will be -1.
	localOutputIndex int32

	// remoteOutputIndex is the output index of this HTLC output in the
	// commitment transaction of the remote node.
	//
	// NOTE: If the output is dust from the PoV of the remote commitment
	// chain, then this value will be -1.
	remoteOutputIndex int32

	// sig is the signature for the second-level HTLC transaction that
	// spends the version of this HTLC on the commitment transaction of the
	// local node. This signature is generated by the remote node and
	// stored by the local node in the case that local node needs to
	// broadcast their commitment transaction.
	sig *btcec.Signature

	// addCommitHeight[Remote|Local] encodes the height of the commitment
	// which included this HTLC on either the remote or local commitment
	// chain. This value is used to determine when an HTLC is fully
	// "locked-in".
	addCommitHeightRemote uint64
	addCommitHeightLocal  uint64

	// removeCommitHeight[Remote|Local] encodes the height of the
	// commitment which removed the parent pointer of this
	// PaymentDescriptor either due to a timeout or a settle. Once both
	// these heights are below the tail of both chains, the log entries can
	// safely be removed.
	removeCommitHeightRemote uint64
	removeCommitHeightLocal  uint64

	// OnionBlob is an opaque blob which is used to complete multi-hop
	// routing.
	//
	// NOTE: Populated only on add payment descriptor entry types.
	OnionBlob []byte

	// ShaOnionBlob is a sha of the onion blob.
	//
	// NOTE: Populated only in payment descriptor with MalfromedFail type.
	ShaOnionBlob [sha256.Size]byte

	// FailReason stores the reason why a particular payment was cancelled.
	//
	// NOTE: Populate only in fail payment descriptor entry types.
	FailReason []byte

	// [our|their|]PkScript are the raw public key scripts that encodes the
	// redemption rules for this particular HTLC. These fields will only be
	// populated iff the EntryType of this PaymentDescriptor is Add.
	// ourPkScript is the ourPkScript from the context of our local
	// commitment chain. theirPkScript is the latest pkScript from the
	// context of the remote commitment chain.
	//
	// NOTE: These values may change within the logs themselves, however,
	// they'll stay consistent within the commitment chain entries
	// themselves.
	ourPkScript        []byte
	ourWitnessScript   []byte
	theirPkScript      []byte
	theirWitnessScript []byte

	// EntryType denotes the exact type of the PaymentDescriptor. In the
	// case of a Timeout, or Settle type, then the Parent field will point
	// into the log to the HTLC being modified.
	EntryType updateType

	// isForwarded denotes if an incoming HTLC has been forwarded to any
	// possible upstream peers in the route.
	isForwarded bool
}