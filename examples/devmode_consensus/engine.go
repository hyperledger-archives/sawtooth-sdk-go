package main

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"math/rand"
	"reflect"
	"strconv"
	"time"

	"github.com/hyperledger/sawtooth-sdk-go/consensus"
	"github.com/hyperledger/sawtooth-sdk-go/logging"
)

var logger *logging.Logger = logging.Get()

const defaultWaitTime = 0

// logGuard is used to keep track of whether
// certain messages have already been logged.
type logGuard struct {
	notReadyToSummarize bool
	notReadyToFinalize  bool
}

// devmodeService pairs a ConsensusService with a logGuard and implements wrapper
// methods around ConsensusService methods.
type devmodeService struct {
	service  consensus.ConsensusService
	logGuard logGuard
}

// NewDevmodeService returns a new devmodeService
func NewDevmodeService(service consensus.ConsensusService) *devmodeService {
	return &devmodeService{
		service: service,
		logGuard: logGuard{
			notReadyToSummarize: false,
			notReadyToFinalize:  false,
		},
	}
}

// getChainHead returns the chain head block, or panics if it is unable.
func (self *devmodeService) getChainHead() consensus.Block {
	logger.Debug("Getting chain head")
	head, err := self.service.GetChainHead()
	if err != nil {
		panic("Failed to get chain head")
	}
	return head
}

// getBlock returns a block for a given block_id. Panics on failure.
func (self *devmodeService) getBlock(block_id consensus.BlockId) consensus.Block {
	logger.Debugf("Getting block %v", block_id)
	ids := make([]consensus.BlockId, 0)
	ids = append(ids, block_id)
	blockIdMap, err := self.service.GetBlocks(ids)
	if err != nil {
		panic("Failed to get block")
	}
	block := blockIdMap[block_id]
	return block
}

// initializeBlock initializes a new block built on the the current block
// head and begins adding batches to it.
func (self *devmodeService) initializeBlock() {
	logger.Debug("Initializing block")
	err := self.service.InitializeBlock(consensus.BLOCK_ID_NULL)
	if err != nil {
		panic("Failed to initialize")
	}
}

// finalizeBlock publishes a block.
func (self *devmodeService) finalizeBlock() consensus.BlockId {
	logger.Debug("Finalizing block")
	var err error

	summary, err := self.service.SummarizeBlock()
	for err != nil && consensus.IsBlockNotReadyError(err) {
		if !self.logGuard.notReadyToSummarize {
			self.logGuard.notReadyToSummarize = true
			logger.Debug("Block not ready to summarize")
		}
		time.Sleep(time.Second * 1)
		summary, err = self.service.SummarizeBlock()
	}
	self.logGuard.notReadyToSummarize = false
	logger.Debug("Block has been summarized successfully")

	newConsensus := createConsensus(summary)
	blockId, err := self.service.FinalizeBlock(newConsensus)
	for err != nil && consensus.IsBlockNotReadyError(err) {
		if !self.logGuard.notReadyToFinalize {
			self.logGuard.notReadyToFinalize = true
			logger.Debug("Block not ready to finalize")
		}
		time.Sleep(time.Second * 1)
		blockId, err = self.service.FinalizeBlock(newConsensus)
	}
	self.logGuard.notReadyToFinalize = false

	logger.Debugf("Block has been finalized successfully: %v", blockId)

	return blockId
}

// checkBlock updates the prioritization of blocks to check. Panics on failure.
func (self *devmodeService) checkBlock(block_id consensus.BlockId) {
	logger.Debugf("Checking block %v", block_id)
	blocks := []consensus.BlockId{block_id}

	err := self.service.CheckBlocks(blocks)
	if err != nil {
		panic("Failed to check block")
	}
}

// failBlock marks this block as invalid from the perspective of consensus. Panics on failure.
func (self *devmodeService) failBlock(block_id consensus.BlockId) {
	logger.Debugf("Failing block %v", block_id)
	err := self.service.FailBlock(block_id)
	if err != nil {
		panic("Failed to fail block")
	}
}

// ignoreBlock signals that this block is no longer being committed. Panics on failure.
func (self *devmodeService) ignoreBlock(block_id consensus.BlockId) {
	logger.Debugf("Ignoring block %v", block_id)
	err := self.service.IgnoreBlock(block_id)
	if err != nil {
		panic("Failed to ignore block")
	}
}

// commitBlock updates the block that should be committed. Panics on failure.
func (self *devmodeService) commitBlock(block_id consensus.BlockId) {
	logger.Debugf("Committing block %v", block_id)
	err := self.service.CommitBlock(block_id)
	if err != nil {
		panic("Failed to commit block")
	}
}

// cancelBlock stops adding batches to the current block and abandons it. Panics on failure.
func (self *devmodeService) cancelBlock() {
	logger.Debug("Canceling block")
	err := self.service.CancelBlock()
	if err != nil && !consensus.IsInvalidStateError(err) {
		panic(fmt.Sprintf("Failed to cancel block: %v", err))
	}
}

// broadcastPublishedBlock broadcasts published block to all connected peers.
// Panics on failure.
func (self *devmodeService) broadcastPublishedBlock(block_id consensus.BlockId) {
	logger.Debugf("Broadcasting published block: %v", block_id)
	err := self.service.Broadcast("published", block_id.AsBytes())
	if err != nil {
		panic("Failed to broadcast published block")
	}
}

// sendBlockReceived sends a consensus message to the block signer. Letting them know
// the block has been recieved. Panics on failure.
func (self devmodeService) sendBlockReceived(block consensus.Block) {
	blockId := block.BlockId()
	err := self.service.SendTo(block.SignerId(), "received", blockId.AsBytes())
	if err != nil {
		panic("Failed to send block received")
	}
}

// sendBlockAck notifies a peer that their block has been acknowledged.
// Panics on failure.
func (self devmodeService) sendBlockAck(sender_id consensus.PeerId, block_id consensus.BlockId) {
	err := self.service.SendTo(sender_id, "ack", block_id.AsBytes())
	if err != nil {
		panic("Failed to send block ack")
	}
}

// calculateWaitTime calculates the time to wait between publishing blocks. This will be a
// random number between the settings sawtooth.consensus.min_wait_time and
// sawtooth.consensus.max_wait_time if max > min, else defaultWaitTime. If
// there is an error parsing those settings, the time will be
// defaultWaitTime.
func (self *devmodeService) calculateWaitTime(chain_head_id consensus.BlockId) time.Duration {
	settings, err := self.service.GetSettings(chain_head_id, []string{"sawtooth.consensus.min_wait_time", "sawtooth.consensus.max_wait_time"})

	waitTime := 0

	if err != nil {

		// get minWaitTime
		minWaitTimeString := settings["sawtooth.consensus.min_wait_time"]
		minWaitTime, err := strconv.Atoi(minWaitTimeString)
		if err != nil {
			minWaitTime = 0
		}

		// get maxWaitTime
		maxWaitTimeString := settings["sawtooth.consensus.max_wait_time"]
		maxWaitTime, err := strconv.Atoi(maxWaitTimeString)
		if err != nil {
			maxWaitTime = 0
		}

		logger.Debugf("Min: %v -- Max: %v", minWaitTime, maxWaitTime)

		if minWaitTime >= maxWaitTime {
			waitTime = defaultWaitTime
		} else {
			// waitTime = value between minWaitTime (inclusive) and maxWaitTime (exclusive)
			rand_range := maxWaitTime - minWaitTime
			waitTime = rand.Intn(rand_range) + minWaitTime
		}

	} else {
		waitTime = defaultWaitTime
	}

	// Convert waitTime from seconds to nanoseconds so we can store it in a time.Duration
	var duration time.Duration = time.Duration(waitTime * 1000000000)

	return duration
}

// A DevmodeEngineImpl implements the ConsensusEngineImpl interface
type DevmodeEngineImpl struct {
	startupState      consensus.StartupState
	service           *devmodeService
	chainHead         consensus.Block
	waitTime          time.Duration
	publishedAtHeight bool
	start             time.Time
}

// Version returns the verson of the DevmodeEngineImpl
func (self *DevmodeEngineImpl) Version() string {
	return "0.1"
}

// Name returns the name of the DevmodeEngineImpl.
// In this case, "Devmode"
func (self *DevmodeEngineImpl) Name() string {
	return "Devmode"
}

// Start is called after the engine is initialized, when a connection to the validator has been
// established.
func (self *DevmodeEngineImpl) Start(startupState consensus.StartupState, service consensus.ConsensusService, updateChan chan consensus.ConsensusUpdate) error {
	self.service = NewDevmodeService(service)
	self.chainHead = startupState.ChainHead()
	self.waitTime = self.service.calculateWaitTime(self.chainHead.BlockId())
	self.publishedAtHeight = false
	self.start = time.Now()

	self.service.initializeBlock()

	for {
		select {
		case n := <-updateChan:
			switch notification := n.(type) {
			case consensus.UpdateShutdown:
				return nil
			case consensus.UpdateBlockNew:
				self.HandleBlockNew(notification.Block)
			case consensus.UpdateBlockValid:
				self.HandleBlockValid(notification.BlockId)
			case consensus.UpdateBlockCommit:
				self.HandleBlockCommit(notification.BlockId)
			case consensus.UpdatePeerMessage:
				self.HandlePeerMessage(notification.PeerMessage, notification.SenderId)
			}
		case <-time.After(time.Millisecond * 10):
			if !self.publishedAtHeight && (time.Now().Sub(self.start) > self.waitTime) {
				logger.Info("Timer expired -- publishing block")
				newBlockId := self.service.finalizeBlock()
				self.publishedAtHeight = true
				self.service.broadcastPublishedBlock(newBlockId)
			}
		}
	}
}

// HandlePeerMessage handles messages recieved from peers.
func (self *DevmodeEngineImpl) HandlePeerMessage(peerMessage consensus.PeerMessage, senderId consensus.PeerId) {
	messageType := peerMessage.Header().MessageType()

	switch messageType {
	case "published":
		logger.Infof("Received block published message from %v: %v", senderId, hex.EncodeToString(peerMessage.Content()))
	case "received":
		logger.Infof("Received block received message from %v: %v", senderId, hex.EncodeToString(peerMessage.Content()))
		self.service.sendBlockAck(senderId, consensus.NewBlockIdFromBytes(peerMessage.Content()))
	case "ack":
		logger.Infof("Received ack message from %v: %v", senderId, hex.EncodeToString(peerMessage.Content()))
	default:
		panic("HandlePeerMessage() recieved an invalid message type")
	}
}

// HandleBlockNew marks an incoming block as valid or invalid, from the perspective of consensus.
func (self *DevmodeEngineImpl) HandleBlockNew(block consensus.Block) {
	logger.Infof("Checking consensus data: %v", block)

	if block.PreviousId() == consensus.BLOCK_ID_NULL {
		logger.Warn("Received genesis block; ignoring")
		return
	}

	if checkConsensus(block) {
		logger.Infof("Passed consensus check: %v", block)
		self.service.checkBlock(block.BlockId())
	} else {
		logger.Infof("Failed consensus check: %v", block)
		self.service.failBlock(block.BlockId())
	}
}

// HandleBlockValid decides whether to commit or ignore a block, based on
// whether it advances the current chainhead or causes a fork to become the
// new chainhead.
func (self *DevmodeEngineImpl) HandleBlockValid(blockId consensus.BlockId) {
	block := self.service.getBlock(blockId)

	self.service.sendBlockReceived(block)

	chainHead := self.service.getChainHead()

	logger.Infof("Choosing between chain heads -- current: %v -- new: %v", chainHead, block)

	// blockBlockIdGreater = if block.BlockId() > chainHead.BlockId()
	blockBlockId := block.BlockId()
	chainHeadId := chainHead.BlockId()
	blockBlockIdGreater := bytes.Compare(blockBlockId.AsBytes(), chainHeadId.AsBytes()) > 0

	// advance the chain if possible
	if block.BlockNum() > chainHead.BlockNum() || (block.BlockNum() == chainHead.BlockNum() && blockBlockIdGreater) {
		logger.Infof("Committing %v", block)
		self.service.commitBlock(blockId)
	} else if block.BlockNum() < chainHead.BlockNum() {
		chainBlock := chainHead
		for {
			chainBlock = self.service.getBlock(chainBlock.PreviousId())
			if chainBlock.BlockNum() == block.BlockNum() {
				break
			}
		}

		// if block.BlockId() > chainBlock.BlockId()
		blockBlockId := block.BlockId()
		chainBlockId := chainBlock.BlockId()
		if bytes.Compare(blockBlockId.AsBytes(), chainBlockId.AsBytes()) > 0 {
			logger.Infof("Switching to new fork %v", block)
			self.service.commitBlock(blockId)
		} else {
			logger.Infof("Ignoring fork %v", block)
			self.service.ignoreBlock(blockId)
		}
	} else {
		logger.Infof("Ignoring %v", block)
		self.service.ignoreBlock(blockId)
	}
}

// HandleBlockCommit abandons the block in progress and starts a new one.
// It should be called after the chain head is updated.
func (self *DevmodeEngineImpl) HandleBlockCommit(newChainHead consensus.BlockId) {
	logger.Infof("Chain head updated to %v, abandoning block in progress", newChainHead)

	self.service.cancelBlock()

	self.waitTime = self.service.calculateWaitTime(newChainHead)
	self.publishedAtHeight = false
	self.start = time.Now()
	self.service.initializeBlock()
}

// checkConsensus ensure consensus is valid via ensuring that block.Payload() is
// equal to block.Summary() with a byte representation of "Devmode" at the begining of it.
func checkConsensus(block consensus.Block) bool {
	return reflect.DeepEqual(block.Payload(), createConsensus(block.Summary()))
}

// createConsensus returns summary with the byte representation of "Devmode" concatenated to the front of it.
func createConsensus(summary []byte) []byte {
	// create a byte slice from the ascii values of a string
	consensusSlice := []byte("Devmode")

	// concatenate the two byte slices
	consensusSlice = append(consensusSlice, summary...)

	return consensusSlice
}
