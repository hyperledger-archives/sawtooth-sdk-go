package consensus

import (
	"encoding/hex"
	"fmt"

	"github.com/golang/protobuf/proto"
	"github.com/hyperledger/sawtooth-sdk-go/protobuf/consensus_pb2"
)

// isSliceOfZeroes returns whether a byte slice contains all zeros.
func isSliceOfZeroes(slice []byte) bool {
	for _, b := range slice {
		if b != 0 {
			return false
		}
	}
	return true
}

// A BlockId is an identifier assigned to every block. BlockIds
// are not encoded in UTF-8.
type BlockId string

const BLOCK_ID_NULL BlockId = ""

// String returns a UTF-8 human-readable hexadecimal representation of a BlockId.
func (self BlockId) String() string {
	return hex.EncodeToString([]byte(self))
}

// AsBytes returns this BlockId as a byte slice.
func (self BlockId) AsBytes() []byte {
	if self == BLOCK_ID_NULL {
		return nil
	}

	return []byte(self)
}

// NewBlockIdFromBytes converts a byte slice and returns it as a BlockId.
func NewBlockIdFromBytes(blockIdBytes []byte) BlockId {
	if isSliceOfZeroes(blockIdBytes) {
		return BLOCK_ID_NULL
	}

	return BlockId(blockIdBytes)
}

// A PeerId is an identifier for a peer.
type PeerId string

const PEER_ID_NULL PeerId = ""

// String returns a UTF-8 human-readable hexadecimal representation of a PeerId.
func (self PeerId) String() string {
	return hex.EncodeToString([]byte(self))
}

// AsBytes returns this PeerId as a byte slice.
func (self PeerId) AsBytes() []byte {
	if self == PEER_ID_NULL {
		return nil
	}

	return []byte(self)
}

// NewPeerIdFromSlice converts a byte slice and returns it as a PeerId.
func NewPeerIdFromSlice(peerIdBytes []byte) PeerId {
	if isSliceOfZeroes(peerIdBytes) {
		return PEER_ID_NULL
	}

	return PeerId(peerIdBytes)
}

// A Block stores all information of a blockchain block.
type Block interface {
	BlockId() BlockId
	PreviousId() BlockId
	SignerId() PeerId
	BlockNum() uint64
	Payload() []byte
	Summary() []byte
}

// a block implements the Block interface.
type block struct {
	blockId    BlockId
	previousId BlockId
	signerId   PeerId
	blockNum   uint64
	payload    []byte
	summary    []byte
}

// newBlockFromProto returns a Block which it assembles from protobuf data.
func newBlockFromProto(blockProto *consensus_pb2.ConsensusBlock) Block {
	return &block{
		blockId:    NewBlockIdFromBytes(blockProto.GetBlockId()),
		previousId: NewBlockIdFromBytes(blockProto.GetPreviousId()),
		signerId:   NewPeerIdFromSlice(blockProto.GetSignerId()),
		blockNum:   blockProto.GetBlockNum(),
		payload:    blockProto.GetPayload(),
		summary:    blockProto.GetSummary(),
	}
}

// BlockId returns the blockId of a block.
func (self *block) BlockId() BlockId {
	return self.blockId
}

// PreviousId returns the previousId of a block.
func (self *block) PreviousId() BlockId {
	return self.previousId
}

// SignerId returns the signerId of a block.
func (self *block) SignerId() PeerId {
	return self.signerId
}

// BlockNum returns the blockNum of a block.
func (self *block) BlockNum() uint64 {
	return self.blockNum
}

// Payload returns the payload of a block.
func (self *block) Payload() []byte {
	return self.payload
}

// Summary returns the summary of a block.
func (self *block) Summary() []byte {
	return self.summary
}

// String returns a UTF-8 human-readable string description of a block.
func (self *block) String() string {
	return fmt.Sprintf("Block(BlockNum: %d, BlockId: %s, PreviousId: %s, SignerId: %s, Payload: %s, Summary: %s)",
		self.BlockNum(),
		self.BlockId(),
		self.PreviousId(),
		self.SignerId(),
		hex.EncodeToString(self.Payload()),
		hex.EncodeToString(self.Summary()),
	)
}

// A PeerInfo is used to retrieve information about a peer.
type PeerInfo interface {
	PeerId() PeerId
}

// A peerInfo stores information about a peer, and implements the PeerInfo interface.
type peerInfo struct {
	peerId PeerId
}

// PeerId returns the peerId of this peerInfo.
func (self *peerInfo) PeerId() PeerId {
	return self.peerId
}

// A PeerMessageHeader allows reading information from a peerMessageHeader.
type PeerMessageHeader interface {
	SignerId() []byte
	ContentSha512() []byte
	MessageType() string
	Name() string
	Version() string
}

// A peerMessageHeader stores information about peer headers, and implements the PeerMessageHeader interface.
type peerMessageHeader struct {
	signerId      []byte
	contentSha512 []byte
	messageType   string
	name          string
	version       string
}

// SignerId returns the signerId of this peerMessageHeader.
func (self *peerMessageHeader) SignerId() []byte {
	return self.signerId
}

// ContentSha512 returns the contentSha512 of this peerMessageHeader.
func (self *peerMessageHeader) ContentSha512() []byte {
	return self.contentSha512
}

// MessageType returns the messageType of this peerMessageHeader.
func (self *peerMessageHeader) MessageType() string {
	return self.messageType
}

// Name returns the name of this peerMessageHeader.
func (self *peerMessageHeader) Name() string {
	return self.name
}

// Version returns the version of this peerMessageHeader.
func (self *peerMessageHeader) Version() string {
	return self.version
}

// A PeerMessage reads information about a peerMessage.
type PeerMessage interface {
	Header() PeerMessageHeader
	HeaderSignature() []byte
	Content() []byte
}

// A peerMessage contains a message from a peer.
type peerMessage struct {
	header          PeerMessageHeader
	headerSignature []byte
	content         []byte
}

// Header returns the header of this peerMessage.
func (self *peerMessage) Header() PeerMessageHeader {
	return self.header
}

// HeaderSignature returns the headerSignature of this peerMessage.
func (self *peerMessage) HeaderSignature() []byte {
	return self.headerSignature
}

// Content returns the content of this peerMessage.
func (self *peerMessage) Content() []byte {
	return self.content
}

// newPeerMessageFromProto returns a PeerMessage from a given protobuf message.
func newPeerMessageFromProto(peerMessageProto *consensus_pb2.ConsensusPeerMessage) PeerMessage {
	headerProto := consensus_pb2.ConsensusPeerMessageHeader{}
	err := proto.Unmarshal(peerMessageProto.GetHeader(), &headerProto)
	if err != nil {
		panic(fmt.Errorf("Error unmarshaling peer message: %v", err))
	}

	header := &peerMessageHeader{
		signerId:      headerProto.GetSignerId(),
		contentSha512: headerProto.GetContentSha512(),
		messageType:   headerProto.GetMessageType(),
		name:          headerProto.GetName(),
		version:       headerProto.GetVersion(),
	}

	message := &peerMessage{
		header:          header,
		headerSignature: peerMessageProto.GetHeaderSignature(),
		content:         peerMessageProto.GetContent(),
	}

	return message
}

// A StartupState returns information on a startupState.
type StartupState interface {
	ChainHead() Block
	Peers() []PeerInfo
	LocalPeerInfo() PeerInfo
}

// A startupState stores information about startup state.
type startupState struct {
	chainHead     Block
	peers         []PeerInfo
	localPeerInfo PeerInfo
}

// ChainHead returns the chainHead of this startupState.
func (self *startupState) ChainHead() Block {
	return self.chainHead
}

// Peers returns the peers of this startupState.
func (self *startupState) Peers() []PeerInfo {
	return self.peers
}

// LocalPeerInfo returns the localPeerInfo of this startupState.
func (self *startupState) LocalPeerInfo() PeerInfo {
	return self.localPeerInfo
}

// newStartupStateFromProtos returns a StartupState from a given protobuf message.
func newStartupStateFromProtos(chainHeadProto *consensus_pb2.ConsensusBlock, peersProto []*consensus_pb2.ConsensusPeerInfo, localPeerInfoProto *consensus_pb2.ConsensusPeerInfo) StartupState {
	peers := make([]PeerInfo, len(peersProto))
	for i, regPeer := range peersProto {
		peers[i] = &peerInfo{
			peerId: NewPeerIdFromSlice(regPeer.GetPeerId()),
		}
	}

	startupState := &startupState{
		chainHead: newBlockFromProto(chainHeadProto),
		peers:     peers,
		localPeerInfo: &peerInfo{
			peerId: NewPeerIdFromSlice(localPeerInfoProto.GetPeerId()),
		},
	}

	return startupState
}

// A Notification contains the type of notification.
type Notification interface {
	GetType() string
}

// A NotificationPeerConnected implements the Notification interface,
// signals the connecting of a peer, and contains it's PeerInfo.
type NotificationPeerConnected struct {
	PeerInfo PeerInfo
}

// GetType returns the notification type of NotificationPeerConnected.
func (self NotificationPeerConnected) GetType() string {
	return "NotificationPeerConnected"
}

// A NotificationPeerDisconnected implements the Notification interface,
// signals the disconnecting of a peer, and contains it's PeerInfo.
type NotificationPeerDisconnected struct {
	PeerInfo PeerInfo
}

// GetType returns the notification type of NotificationPeerDisconnected.
func (self NotificationPeerDisconnected) GetType() string {
	return "NotificationPeerDisconnected"
}

// A NotificationPeerMessage implements the Notification interface,
// and signals an incoming message from a peer.
type NotificationPeerMessage struct {
	PeerMessage PeerMessage
	SenderId    PeerId
}

// GetType returns the notification type of NotificationPeerMessage.
func (self NotificationPeerMessage) GetType() string {
	return "NotificationPeerMessage"
}

// A NotificationBlockNew implements the Notification interface,
// and signals the arrival of a new block.
type NotificationBlockNew struct {
	Block Block
}

// GetType returns the notification type of NotificationBlockNew.
func (self NotificationBlockNew) GetType() string {
	return "NotificationBlockNew"
}

// A NotificationBlockValid implements the Notification interface,
// and signals that a block is valid.
type NotificationBlockValid struct {
	BlockId BlockId
}

// GetType returns the notification type of NotificationBlockValid.
func (self NotificationBlockValid) GetType() string {
	return "NotificationBlockValid"
}

// A NotificationBlockInvalid implements the Notification interface,
// and signals that a block is invalid.
type NotificationBlockInvalid struct {
	BlockId BlockId
}

// GetType returns the notification type of NotificationBlockInvalid.
func (self NotificationBlockInvalid) GetType() string {
	return "NotificationBlockInvalid"
}

// A NotificationBlockCommit implements the Notification interface,
// and signals that a block has been commited.
type NotificationBlockCommit struct {
	BlockId BlockId
}

// GetType returns the notification type of NotificationBlockCommit.
func (self NotificationBlockCommit) GetType() string {
	return "NotificationBlockCommit"
}

// A NotificationShutdown implements the Notification interface,
// and signals that we should shutdown.
type NotificationShutdown struct{}

// GetType returns the notification type of NotificationShutdown.
func (self NotificationShutdown) GetType() string {
	return "NotificationShutdown"
}
