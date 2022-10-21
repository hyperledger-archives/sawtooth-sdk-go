/**
 * Copyright 2022 Grid 7, LLC (DBA Taekion)
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 * ------------------------------------------------------------------------------
 */

package consensus

import "fmt"

type EncodingError struct {
	Msg string
}

func (self *EncodingError) Error() string {
	return fmt.Sprintf("EncodingError: %s", self.Msg)
}

// IsEncodingError returns whether the error is an EncodingError.
func IsEncodingError(err error) bool {
	_, ok := err.(*EncodingError)
	return ok
}

type SendError struct {
	Msg string
}

func (self *SendError) Error() string {
	return fmt.Sprintf("SendError: %s", self.Msg)
}

// IsSendError returns whether the error is a SendError.
func IsSendError(err error) bool {
	_, ok := err.(*SendError)
	return ok
}

type ReceiveError struct {
	Msg string
}

func (self *ReceiveError) Error() string {
	return fmt.Sprintf("ReceiveError: %s", self.Msg)
}

// IsReceiveError returns whether the error is a ReceiveError.
func IsReceiveError(err error) bool {
	_, ok := err.(*ReceiveError)
	return ok
}

type InvalidStateError struct {
	Msg string
}

func (self *InvalidStateError) Error() string {
	return fmt.Sprintf("InvalidState: %s", self.Msg)
}

// IsInvalidStateError returns whether the error is a InvalidStateError.
func IsInvalidStateError(err error) bool {
	_, ok := err.(*InvalidStateError)
	return ok
}

type UnknownBlockError struct {
	Msg string
}

func (self *UnknownBlockError) Error() string {
	return fmt.Sprintf("UnknownBlock: %s", self.Msg)
}

// IsUnknownBlockError returns whether the error is a UnknownBlockError.
func IsUnknownBlockError(err error) bool {
	_, ok := err.(*UnknownBlockError)
	return ok
}

type UnknownPeerError struct {
	Msg string
}

func (self *UnknownPeerError) Error() string {
	return fmt.Sprintf("UnknownPeer: %s", self.Msg)
}

// IsUnknownPeerError returns whether the error is a UnknownPeerError.
func IsUnknownPeerError(err error) bool {
	_, ok := err.(*UnknownPeerError)
	return ok
}

type NoChainHeadError struct{}

func (self *NoChainHeadError) Error() string {
	return fmt.Sprint("NoChainHead")
}

// IsNoChainHeadError returns whether the error is a NoChainHeadError.
func IsNoChainHeadError(err error) bool {
	_, ok := err.(*NoChainHeadError)
	return ok
}

type BlockNotReadyError struct{}

func (self *BlockNotReadyError) Error() string {
	return fmt.Sprint("BlockNotReady")
}

// IsBlockNotReadyError returns whether the error is a BlockNotReadyError.
func IsBlockNotReadyError(err error) bool {
	_, ok := err.(*BlockNotReadyError)
	return ok
}
