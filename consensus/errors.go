package consensus

import "fmt"

type EncodingError struct {
	Msg string
}

func (self *EncodingError) Error() string {
	return fmt.Sprintf("EncodingError: %s", self.Msg)
}

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

func IsUnknownPeerError(err error) bool {
	_, ok := err.(*UnknownPeerError)
	return ok
}

type NoChainHeadError struct{}

func (self *NoChainHeadError) Error() string {
	return fmt.Sprint("NoChainHead")
}

func IsNoChainHeadError(err error) bool {
	_, ok := err.(*NoChainHeadError)
	return ok
}

type BlockNotReadyError struct{}

func (self *BlockNotReadyError) Error() string {
	return fmt.Sprint("BlockNotReady")
}

func IsBlockNotReadyError(err error) bool {
	_, ok := err.(*BlockNotReadyError)
	return ok
}
