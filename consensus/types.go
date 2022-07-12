package consensus

import (
	"encoding/hex"
	"fmt"
	"github.com/golang/protobuf/proto"
	"github.com/hyperledger/sawtooth-sdk-go/protobuf/consensus_pb2"
)

func isSliceOfZeroes(slice []byte) bool {
	for _, b := range slice {
		if b != 0 {
			return false
		}
	}
	return true
}

type BlockId string

const BLOCK_ID_NULL BlockId = ""

func (self BlockId) String() string {
	return hex.EncodeToString([]byte(self))
}

func (self BlockId) AsBytes() []byte {
	if self == BLOCK_ID_NULL {
		return nil
	}

	return []byte(self)
}

func NewBlockIdFromBytes(blockIdBytes []byte) BlockId {
	if isSliceOfZeroes(blockIdBytes) {
		return BLOCK_ID_NULL
	}

	return BlockId(blockIdBytes)
}

type PeerId string

const PEER_ID_NULL PeerId = ""

func (self PeerId) String() string {
	return hex.EncodeToString([]byte(self))
}

func (self PeerId) AsBytes() []byte {
	if self == PEER_ID_NULL {
		return nil
	}

	return []byte(self)
}

func NewPeerIdFromSlice(peerIdBytes []byte) PeerId {
	if isSliceOfZeroes(peerIdBytes) {
		return PEER_ID_NULL
	}

	return PeerId(peerIdBytes)
}

type Block interface {
	BlockId() BlockId
	PreviousId() BlockId
	SignerId() PeerId
	BlockNum() uint64
	Payload() []byte
	Summary() []byte
}

type block struct {
	blockId    BlockId
	previousId BlockId
	signerId   PeerId
	blockNum   uint64
	payload    []byte
	summary    []byte
}

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

func (self *block) BlockId() BlockId {
	return self.blockId
}

func (self *block) PreviousId() BlockId {
	return self.previousId
}

func (self *block) SignerId() PeerId {
	return self.signerId
}

func (self *block) BlockNum() uint64 {
	return self.blockNum
}

func (self *block) Payload() []byte {
	return self.payload
}

func (self *block) Summary() []byte {
	return self.summary
}

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

type PeerInfo interface {
	PeerId() PeerId
}

type peerInfo struct {
	peerId PeerId
}

func (self *peerInfo) PeerId() PeerId {
	return self.peerId
}

type PeerMessageHeader interface {
	SignerId() []byte
	ContentSha512() []byte
	MessageType() string
	Name() string
	Version() string
}

type peerMessageHeader struct {
	signerId      []byte
	contentSha512 []byte
	messageType   string
	name          string
	version       string
}

func (self *peerMessageHeader) SignerId() []byte {
	return self.signerId
}

func (self *peerMessageHeader) ContentSha512() []byte {
	return self.contentSha512
}

func (self *peerMessageHeader) MessageType() string {
	return self.messageType
}

func (self *peerMessageHeader) Name() string {
	return self.name
}

func (self *peerMessageHeader) Version() string {
	return self.version
}

type PeerMessage interface {
	Header() PeerMessageHeader
	HeaderSignature() []byte
	Content() []byte
}

type peerMessage struct {
	header          PeerMessageHeader
	headerSignature []byte
	content         []byte
}

func (self *peerMessage) Header() PeerMessageHeader {
	return self.header
}

func (self *peerMessage) HeaderSignature() []byte {
	return self.headerSignature
}

func (self *peerMessage) Content() []byte {
	return self.content
}

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

type StartupState interface {
	ChainHead() Block
	Peers() []PeerInfo
	LocalPeerInfo() PeerInfo
}

type startupState struct {
	chainHead     Block
	peers         []PeerInfo
	localPeerInfo PeerInfo
}

func (self *startupState) ChainHead() Block {
	return self.chainHead
}

func (self *startupState) Peers() []PeerInfo {
	return self.peers
}

func (self *startupState) LocalPeerInfo() PeerInfo {
	return self.localPeerInfo
}

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

type Notification interface {
	GetType() string
}

type NotificationPeerConnected struct {
	PeerInfo PeerInfo
}

func (self NotificationPeerConnected) GetType() string {
	return "NotificationPeerConnected"
}

type NotificationPeerDisconnected struct {
	PeerInfo PeerInfo
}

func (self NotificationPeerDisconnected) GetType() string {
	return "NotificationPeerDisconnected"
}

type NotificationPeerMessage struct {
	PeerMessage PeerMessage
	SenderId    PeerId
}

func (self NotificationPeerMessage) GetType() string {
	return "NotificationPeerMessage"
}

type NotificationBlockNew struct {
	Block Block
}

func (self NotificationBlockNew) GetType() string {
	return "NotificationBlockNew"
}

type NotificationBlockValid struct {
	BlockId BlockId
}

func (self NotificationBlockValid) GetType() string {
	return "NotificationBlockValid"
}

type NotificationBlockInvalid struct {
	BlockId BlockId
}

func (self NotificationBlockInvalid) GetType() string {
	return "NotificationBlockInvalid"
}

type NotificationBlockCommit struct {
	BlockId BlockId
}

func (self NotificationBlockCommit) GetType() string {
	return "NotificationBlockCommit"
}

type NotificationShutdown struct{}

func (self NotificationShutdown) GetType() string {
	return "NotificationShutdown"
}
