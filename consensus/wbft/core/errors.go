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
// This file is derived from quorum/consensus/istanbul/qbft/core/errors.go (2024.07.25).
// Modified and improved for the wemix development.

package core

import "errors"

var (
	// errNotFromProposer is returned when received message is supposed to be from
	// proposer.
	errNotFromProposer = errors.New("message does not come from proposer")
	// errFutureMessage is returned when current view is earlier than the
	// view of the received message.
	errFutureMessage = errors.New("future message")
	// errOldMessage is returned when the received message's view is earlier
	// than current view.
	errOldMessage = errors.New("old message")
	// errInvalidMessage is returned when the message is malformed.
	errInvalidMessage = errors.New("invalid message")
	// errInvalidSeal is returned when the signed seal is not matched with message
	errInvalidSeal = errors.New("invalid seal")
	// errInvalidSigner is returned when the message is signed by a validator different than message sender
	errInvalidSigner = errors.New("message not signed by the sender")
	// errInvalidPreparedBlock is returned when prepared block is not validated in round change messages
	errInvalidPreparedBlock = errors.New("invalid prepared block in round change messages")
	// errExtraSealMessage is returned when messages which will be added to next block's prevSeals comes.
	errExtraSealMessage = errors.New("extra seal message")
	// errInvalidExtraSealMessage is returned when message is not appropriate to be added to extraSeals
	errInvalidExtraSealMessage = errors.New("invalid extra seal message")
	errCurrentIsNil            = errors.New("current is nil")

	// errFutureMessage is returned when the received message has a view
	// significantly ahead of the current view (in sequence or round).
	errFutureViewTooFar = errors.New("future view too far ahead: sequence or round difference too large")
)
