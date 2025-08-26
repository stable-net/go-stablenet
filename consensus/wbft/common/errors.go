// Copyright 2017 The go-ethereum Authors
// Copyright 2024 The go-wemix-wbft Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.
//
// This file is derived from quorum/consensus/istanbul/common/errors.go (2024.07.25).
// Modified and improved for the wemix development.

package wbftcommon

import "errors"

var (
	// ErrInvalidProposal is returned when a prposal is malformed.
	ErrInvalidProposal = errors.New("invalid proposal")

	// ErrInvalidSignature is returned when given signature is not signed by given
	// address.
	ErrInvalidSignature = errors.New("invalid signature")

	// ErrUnknownBlock is returned when the list of validators is requested for a block
	// that is not part of the local blockchain.
	ErrUnknownBlock = errors.New("unknown block")

	// ErrUnauthorized is returned if a header is signed by a non authorized entity.
	ErrUnauthorized = errors.New("unauthorized")

	// ErrInvalidDifficulty is returned if the difficulty of a block is not 1
	ErrInvalidDifficulty = errors.New("invalid difficulty")

	// ErrInvalidExtraDataFormat is returned when the extra data format is incorrect
	ErrInvalidExtraDataFormat = errors.New("invalid extra data format")

	// ErrInvalidMixDigest is returned if a block's mix digest is not Istanbul digest.
	ErrInvalidMixDigest = errors.New("invalid Istanbul mix digest")

	// ErrInvalidNonce is returned if a block's nonce is invalid
	ErrInvalidNonce = errors.New("invalid nonce")

	// ErrInvalidUncleHash is returned if a block contains an non-empty uncle list.
	ErrInvalidUncleHash = errors.New("non empty uncle hash")

	// ErrBlacklistedHash is returned if a block to import is on the blacklist.
	ErrBlacklistedHash = errors.New("blacklisted hash") //

	// ErrInconsistentValidatorSet is returned if the validator set is inconsistent
	// ErrInconsistentValidatorSet = errors.New("non empty uncle hash")
	// ErrInvalidTimestamp is returned if the timestamp of a block is lower than the previous block's timestamp + the minimum block period.
	ErrInvalidTimestamp = errors.New("invalid timestamp")

	// ErrInvalidVotingChain is returned if an authorization list is attempted to
	// be modified via out-of-range or non-contiguous headers.
	ErrInvalidVotingChain = errors.New("invalid voting chain")

	// ErrInvalidVote is returned if a nonce value is something else that the two
	// allowed constants of 0x00..0 or 0xff..f.
	ErrInvalidVote = errors.New("vote nonce not 0x00..0 or 0xff..f")

	// ErrInvalidPreparedSeals is returned if the prepared seal is not signed by any of parent validators.
	ErrInvalidPreparedSeals = errors.New("invalid prepared seals")

	// ErrInvalidPrevPreparedSeals is returned if the prev prepared seal is not signed by any of parent validators or not enough seals.
	ErrInvalidPrevPreparedSeals = errors.New("invalid prev prepared seals")

	// ErrEmptyPreparedSeals is returned if the field of prepared seals is zero.
	ErrEmptyPreparedSeals = errors.New("zero prepared seals")

	// ErrEmptyPrevPreparedSeals is returned if the field of prepared seals is zero.
	ErrEmptyPrevPreparedSeals = errors.New("zero prev prepared seals")

	// ErrInvalidCommittedSeals is returned if the committed seal is not signed by any of parent validators.
	ErrInvalidCommittedSeals = errors.New("invalid committed seals")

	// ErrInvalidPrevCommittedSeals is returned if the prev committed seal is not signed by any of parent validators or not enough seals.
	ErrInvalidPrevCommittedSeals = errors.New("invalid prev committed seals")

	// ErrEmptyCommittedSeals is returned if the field of committed seals is zero.
	ErrEmptyCommittedSeals = errors.New("zero committed seals")

	// ErrEmptyPrevCommittedSeals is returned if the field of committed seals is zero.
	ErrEmptyPrevCommittedSeals = errors.New("zero prev committed seals")

	// ErrInvalidSeal is returned if the seal is not signed by any of validators.
	ErrInvalidSeal = errors.New("invalid seal")

	// ErrEmptySeals is returned if the field of seals is zero.
	ErrEmptySeals = errors.New("zero seals")

	// ErrMismatchTxhashes is returned if the TxHash in header is mismatch.
	ErrMismatchTxhashes = errors.New("mismatch transactions hashes")

	// ErrInconsistentSubject is returned when received subject is different from
	// current subject.
	ErrInconsistentSubject = errors.New("inconsistent subjects")

	// ErrNotFromProposer is returned when received message is supposed to be from
	// proposer.
	ErrNotFromProposer = errors.New("message does not come from proposer")

	// ErrIgnored is returned when a message was ignored.
	ErrIgnored = errors.New("message is ignored")

	// ErrOldMessage is returned when the received message's view is earlier
	// than current view.
	ErrOldMessage = errors.New("old message")

	// ErrInvalidMessage is returned when the message is malformed.
	ErrInvalidMessage = errors.New("invalid message")

	// ErrFailedDecodePreprepare is returned when the PRE-PREPARE message is malformed.
	ErrFailedDecodePreprepare = errors.New("failed to decode PRE-PREPARE message")

	// ErrFailedDecodePrepare is returned when the PREPARE message is malformed.
	ErrFailedDecodePrepare = errors.New("failed to decode PREPARE message")

	// ErrFailedDecodeCommit is returned when the COMMIT message is malformed.
	ErrFailedDecodeCommit = errors.New("failed to decode COMMIT message")

	// ErrFailedDecodeRoundChange is returned when the COMMIT message is malformed.
	ErrFailedDecodeRoundChange = errors.New("failed to decode ROUND-CHANGE message")

	// ErrFailedDecodeMessageSet is returned when the message set is malformed.
	// ErrFailedDecodeMessageSet = errors.New("failed to decode message set")
	// ErrInvalidSigner is returned when the message is signed by a validator different than message sender
	ErrInvalidSigner = errors.New("message not signed by the sender")

	ErrInvalidGenesis = errors.New("genesis must only specify single validator mode for block zero")

	ErrInvalidSpecificCall = errors.New("invalid method name for engine specific function")

	ErrIsNotWBFTBlock = errors.New("block is not a wbft block")
)
