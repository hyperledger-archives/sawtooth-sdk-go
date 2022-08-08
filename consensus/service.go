package consensus

import (
	"fmt"

	"github.com/golang/protobuf/proto"
	"github.com/hyperledger/sawtooth-sdk-go/messaging"
	"github.com/hyperledger/sawtooth-sdk-go/protobuf/consensus_pb2"
	"github.com/hyperledger/sawtooth-sdk-go/protobuf/validator_pb2"
	zmq "github.com/pebbe/zmq4"
)

// A ZmqService implements the ConsensusService interface, which provides
// methods that allow the consensus engine to issue commands and requests.
type ZmqService struct {
	context *zmq.Context
	uri     string
	// Connection is for outgoing requests from this service.
	// Responses come back over the same connection.
	connection messaging.Connection
	corrIds    map[string]interface{}
}

// NewZmqService returns a new ZmqService.
func NewZmqService(context *zmq.Context, uri string) *ZmqService {
	return &ZmqService{
		context: context,
		uri:     uri,
		corrIds: make(map[string]interface{}),
	}
}

// Start establishes a new ZMQ connection.
func (self *ZmqService) Start() {
	connection, err := messaging.NewConnection(self.context, zmq.PAIR, self.uri, false)
	if err != nil {
		logger.Errorf("Failed to connect to main thread via service socket: %v", err)
		return
	}
	self.connection = connection
}

// Shutdown closes the active ZMQ connection.
func (self *ZmqService) Shutdown() {
	self.connection.Close()
}

// SendTo sends a consensus message to a specific, connected peer.
func (self *ZmqService) SendTo(peer PeerId, messageType string, payload []byte) error {
	// Prepare the request message
	request := consensus_pb2.ConsensusSendToRequest{
		Content:     payload,
		MessageType: messageType,
		ReceiverId:  peer.AsBytes(),
	}
	response := consensus_pb2.ConsensusSendToResponse{}
	requestType := validator_pb2.Message_CONSENSUS_SEND_TO_REQUEST
	responseType := validator_pb2.Message_CONSENSUS_SEND_TO_RESPONSE

	err := self.serviceRpc(&request, requestType, &response, responseType)
	if err != nil {
		return err
	}

	// Check for OK status
	if response.GetStatus() != consensus_pb2.ConsensusSendToResponse_OK {
		return &ReceiveError{
			Msg: fmt.Sprintf("Failed with status: %v", response.GetStatus()),
		}
	}

	return nil
}

// Broadcast broadcasts a message to all connected peers.
func (self *ZmqService) Broadcast(messageType string, payload []byte) error {
	// Prepare the request message
	request := consensus_pb2.ConsensusBroadcastRequest{
		Content:     payload,
		MessageType: messageType,
	}
	response := consensus_pb2.ConsensusBroadcastResponse{}
	requestType := validator_pb2.Message_CONSENSUS_BROADCAST_REQUEST
	responseType := validator_pb2.Message_CONSENSUS_BROADCAST_RESPONSE

	err := self.serviceRpc(&request, requestType, &response, responseType)
	if err != nil {
		return err
	}

	// Check for OK status
	if response.GetStatus() != consensus_pb2.ConsensusBroadcastResponse_OK {
		return &ReceiveError{
			Msg: fmt.Sprintf("Failed with status: %v", response.GetStatus()),
		}
	}

	return nil
}

// InitializeBlock initializes a new block built on the block with
// the given previous id and begins adding batches to it. If no
// previous id is specified, the current head will be used.
func (self *ZmqService) InitializeBlock(previousId BlockId) error {
	// Prepare the request message
	request := consensus_pb2.ConsensusInitializeBlockRequest{
		PreviousId: previousId.AsBytes(),
	}
	response := consensus_pb2.ConsensusInitializeBlockResponse{}
	requestType := validator_pb2.Message_CONSENSUS_INITIALIZE_BLOCK_REQUEST
	responseType := validator_pb2.Message_CONSENSUS_INITIALIZE_BLOCK_RESPONSE

	err := self.serviceRpc(&request, requestType, &response, responseType)
	if err != nil {
		return err
	}

	if response.GetStatus() == consensus_pb2.ConsensusInitializeBlockResponse_INVALID_STATE {
		return &InvalidStateError{
			Msg: "Cannot initialize block in current state",
		}
	} else if response.GetStatus() == consensus_pb2.ConsensusInitializeBlockResponse_UNKNOWN_BLOCK {
		return &UnknownBlockError{
			Msg: "Block not found",
		}
	} else if response.GetStatus() != consensus_pb2.ConsensusInitializeBlockResponse_OK {
		return &ReceiveError{
			Msg: fmt.Sprintf("Failed with status: %v", response.GetStatus()),
		}
	}

	return nil
}

// SummarizeBlock stops adding batches to the current block
// and returns a summary of its contents.
func (self *ZmqService) SummarizeBlock() ([]byte, error) {
	request := consensus_pb2.ConsensusSummarizeBlockRequest{}
	response := consensus_pb2.ConsensusSummarizeBlockResponse{}
	requestType := validator_pb2.Message_CONSENSUS_SUMMARIZE_BLOCK_REQUEST
	responseType := validator_pb2.Message_CONSENSUS_SUMMARIZE_BLOCK_RESPONSE

	err := self.serviceRpc(&request, requestType, &response, responseType)
	if err != nil {
		return nil, err
	}

	if response.GetStatus() == consensus_pb2.ConsensusSummarizeBlockResponse_INVALID_STATE {
		return nil, &InvalidStateError{
			Msg: "Cannot summarize block in current state",
		}
	} else if response.GetStatus() == consensus_pb2.ConsensusSummarizeBlockResponse_BLOCK_NOT_READY {
		return nil, &BlockNotReadyError{}
	} else if response.GetStatus() != consensus_pb2.ConsensusSummarizeBlockResponse_OK {
		return nil, &ReceiveError{
			Msg: fmt.Sprintf("Failed with status: %v", response.GetStatus()),
		}
	}

	return response.GetSummary(), nil
}

// FinalizeBlock inserts the given consensus data into the block and
// signs it. If this call is successful, the consensus engine will
// receive the block afterwards.
func (self *ZmqService) FinalizeBlock(data []byte) (BlockId, error) {
	request := consensus_pb2.ConsensusFinalizeBlockRequest{
		Data: data,
	}
	response := consensus_pb2.ConsensusFinalizeBlockResponse{}
	requestType := validator_pb2.Message_CONSENSUS_FINALIZE_BLOCK_REQUEST
	responseType := validator_pb2.Message_CONSENSUS_FINALIZE_BLOCK_RESPONSE

	err := self.serviceRpc(&request, requestType, &response, responseType)
	if err != nil {
		return BLOCK_ID_NULL, err
	}

	if response.GetStatus() == consensus_pb2.ConsensusFinalizeBlockResponse_INVALID_STATE {
		return BLOCK_ID_NULL, &InvalidStateError{
			Msg: "Cannot finalize block in current state",
		}
	} else if response.GetStatus() == consensus_pb2.ConsensusFinalizeBlockResponse_BLOCK_NOT_READY {
		return BLOCK_ID_NULL, &BlockNotReadyError{}
	} else if response.GetStatus() != consensus_pb2.ConsensusFinalizeBlockResponse_OK {
		return BLOCK_ID_NULL, &ReceiveError{
			Msg: fmt.Sprintf("Failed with status: %v", response.GetStatus()),
		}
	}

	return NewBlockIdFromBytes(response.GetBlockId()), nil
}

// CancelBlock stops adding batches to the current block and abandons it.
func (self *ZmqService) CancelBlock() error {
	request := consensus_pb2.ConsensusCancelBlockRequest{}
	response := consensus_pb2.ConsensusCancelBlockResponse{}
	requestType := validator_pb2.Message_CONSENSUS_CANCEL_BLOCK_REQUEST
	responseType := validator_pb2.Message_CONSENSUS_CANCEL_BLOCK_RESPONSE

	err := self.serviceRpc(&request, requestType, &response, responseType)
	if err != nil {
		return err
	}

	// Check for OK status
	if response.GetStatus() == consensus_pb2.ConsensusCancelBlockResponse_INVALID_STATE {
		return &InvalidStateError{
			Msg: "Cannot cancel block in current state",
		}
	} else if response.GetStatus() != consensus_pb2.ConsensusCancelBlockResponse_OK {
		return &ReceiveError{
			Msg: fmt.Sprintf("Failed with status: %v", response.GetStatus()),
		}
	}

	return nil
}

// CheckBlocks updates the prioritization of blocks to check.
func (self *ZmqService) CheckBlocks(blockIds []BlockId) error {
	logger.Debug(blockIds)
	blockIdsRequest := make([][]byte, len(blockIds))
	for i, blockId := range blockIds {
		blockIdsRequest[i] = blockId.AsBytes()
	}

	request := consensus_pb2.ConsensusCheckBlocksRequest{
		BlockIds: blockIdsRequest,
	}
	response := consensus_pb2.ConsensusCheckBlocksResponse{}
	requestType := validator_pb2.Message_CONSENSUS_CHECK_BLOCKS_REQUEST
	responseType := validator_pb2.Message_CONSENSUS_CHECK_BLOCKS_RESPONSE

	err := self.serviceRpc(&request, requestType, &response, responseType)
	if err != nil {
		return err
	}

	if response.GetStatus() == consensus_pb2.ConsensusCheckBlocksResponse_UNKNOWN_BLOCK {
		return &UnknownBlockError{
			Msg: "Block not found",
		}
	} else if response.GetStatus() != consensus_pb2.ConsensusCheckBlocksResponse_OK {
		return &ReceiveError{
			Msg: fmt.Sprintf("Failed with status: %v", response.GetStatus()),
		}
	}

	return nil
}

// CommitBlock updates the block that should be committed.
func (self *ZmqService) CommitBlock(blockId BlockId) error {
	request := consensus_pb2.ConsensusCommitBlockRequest{
		BlockId: blockId.AsBytes(),
	}
	response := consensus_pb2.ConsensusCommitBlockResponse{}
	requestType := validator_pb2.Message_CONSENSUS_COMMIT_BLOCK_REQUEST
	responseType := validator_pb2.Message_CONSENSUS_COMMIT_BLOCK_RESPONSE

	err := self.serviceRpc(&request, requestType, &response, responseType)
	if err != nil {
		return err
	}

	if response.GetStatus() == consensus_pb2.ConsensusCommitBlockResponse_UNKNOWN_BLOCK {
		return &UnknownBlockError{
			Msg: "Block not found",
		}
	} else if response.GetStatus() != consensus_pb2.ConsensusCommitBlockResponse_OK {
		return &ReceiveError{
			Msg: fmt.Sprintf("Failed with status: %v", response.GetStatus()),
		}
	}

	return nil
}

// IgnoreBlock signals that this block is no longer being committed.
func (self *ZmqService) IgnoreBlock(blockId BlockId) error {
	request := consensus_pb2.ConsensusIgnoreBlockRequest{
		BlockId: blockId.AsBytes(),
	}
	response := consensus_pb2.ConsensusIgnoreBlockResponse{}
	requestType := validator_pb2.Message_CONSENSUS_IGNORE_BLOCK_REQUEST
	responseType := validator_pb2.Message_CONSENSUS_IGNORE_BLOCK_RESPONSE

	err := self.serviceRpc(&request, requestType, &response, responseType)
	if err != nil {
		return err
	}

	if response.GetStatus() == consensus_pb2.ConsensusIgnoreBlockResponse_UNKNOWN_BLOCK {
		return &UnknownBlockError{
			Msg: "Block not found",
		}
	} else if response.GetStatus() != consensus_pb2.ConsensusIgnoreBlockResponse_OK {
		return &ReceiveError{
			Msg: fmt.Sprintf("Failed with status: %v", response.GetStatus()),
		}
	}

	return nil
}

// FailBlock marks this block as invalid from the perspective of consensus.
func (self *ZmqService) FailBlock(blockId BlockId) error {
	request := consensus_pb2.ConsensusFailBlockRequest{
		BlockId: blockId.AsBytes(),
	}
	response := consensus_pb2.ConsensusFailBlockResponse{}
	requestType := validator_pb2.Message_CONSENSUS_FAIL_BLOCK_REQUEST
	responseType := validator_pb2.Message_CONSENSUS_FAIL_BLOCK_RESPONSE

	err := self.serviceRpc(&request, requestType, &response, responseType)
	if err != nil {
		return err
	}

	if response.GetStatus() == consensus_pb2.ConsensusFailBlockResponse_UNKNOWN_BLOCK {
		return &UnknownBlockError{
			Msg: "Block not found",
		}
	} else if response.GetStatus() != consensus_pb2.ConsensusFailBlockResponse_OK {
		return &ReceiveError{
			Msg: fmt.Sprintf("Failed with status: %v", response.GetStatus()),
		}
	}

	return nil
}

// GetBlocks returns consensus-related information about blocks.
func (self *ZmqService) GetBlocks(blockIds []BlockId) (map[BlockId]Block, error) {
	blockIdsRequest := make([][]byte, len(blockIds))
	for i, blockId := range blockIds {
		blockIdsRequest[i] = blockId.AsBytes()
	}

	request := consensus_pb2.ConsensusBlocksGetRequest{
		BlockIds: blockIdsRequest,
	}
	response := consensus_pb2.ConsensusBlocksGetResponse{}
	requestType := validator_pb2.Message_CONSENSUS_BLOCKS_GET_REQUEST
	responseType := validator_pb2.Message_CONSENSUS_BLOCKS_GET_RESPONSE

	err := self.serviceRpc(&request, requestType, &response, responseType)
	if err != nil {
		return nil, err
	}

	if response.GetStatus() == consensus_pb2.ConsensusBlocksGetResponse_UNKNOWN_BLOCK {
		return nil, &UnknownBlockError{
			Msg: "Block not found",
		}
	} else if response.GetStatus() != consensus_pb2.ConsensusBlocksGetResponse_OK {
		return nil, &ReceiveError{
			Msg: fmt.Sprintf("Failed with status: %v", response.GetStatus()),
		}
	}

	results := make(map[BlockId]Block, len(blockIds))
	for _, blockProto := range response.GetBlocks() {
		results[NewBlockIdFromBytes(blockProto.GetBlockId())] = newBlockFromProto(blockProto)
	}
	return results, nil
}

// GetChainHead gets the chain head block.
func (self *ZmqService) GetChainHead() (Block, error) {
	request := consensus_pb2.ConsensusChainHeadGetRequest{}
	response := consensus_pb2.ConsensusChainHeadGetResponse{}
	requestType := validator_pb2.Message_CONSENSUS_CHAIN_HEAD_GET_REQUEST
	responseType := validator_pb2.Message_CONSENSUS_CHAIN_HEAD_GET_RESPONSE

	err := self.serviceRpc(&request, requestType, &response, responseType)
	if err != nil {
		return nil, err
	}

	if response.GetStatus() == consensus_pb2.ConsensusChainHeadGetResponse_NO_CHAIN_HEAD {
		return nil, &NoChainHeadError{}
	} else if response.GetStatus() != consensus_pb2.ConsensusChainHeadGetResponse_OK {
		return nil, &ReceiveError{
			Msg: fmt.Sprintf("Failed with status: %v", response.GetStatus()),
		}
	}

	return newBlockFromProto(response.GetBlock()), nil
}

// GetSettings returns the value of settings as of the given block.
func (self *ZmqService) GetSettings(blockId BlockId, keys []string) (map[string]string, error) {
	request := consensus_pb2.ConsensusSettingsGetRequest{
		BlockId: blockId.AsBytes(),
		Keys:    keys,
	}
	response := consensus_pb2.ConsensusSettingsGetResponse{}
	requestType := validator_pb2.Message_CONSENSUS_SETTINGS_GET_REQUEST
	responseType := validator_pb2.Message_CONSENSUS_SETTINGS_GET_RESPONSE

	err := self.serviceRpc(&request, requestType, &response, responseType)
	if err != nil {
		return nil, err
	}

	if response.GetStatus() == consensus_pb2.ConsensusSettingsGetResponse_UNKNOWN_BLOCK {
		return nil, &UnknownBlockError{
			Msg: "Block not found",
		}
	} else if response.GetStatus() != consensus_pb2.ConsensusSettingsGetResponse_OK {
		return nil, &ReceiveError{
			Msg: fmt.Sprintf("Failed with status: %v", response.GetStatus()),
		}
	}

	result := make(map[string]string, len(response.GetEntries()))
	for _, entry := range response.GetEntries() {
		result[entry.GetKey()] = entry.GetValue()
	}
	return result, nil
}

// GetState returns values in state as of the given block.
func (self *ZmqService) GetState(blockId BlockId, addresses []string) (map[string][]byte, error) {
	request := consensus_pb2.ConsensusStateGetRequest{
		BlockId:   blockId.AsBytes(),
		Addresses: addresses,
	}
	response := consensus_pb2.ConsensusStateGetResponse{}
	requestType := validator_pb2.Message_CONSENSUS_STATE_GET_REQUEST
	responseType := validator_pb2.Message_CONSENSUS_STATE_GET_RESPONSE

	err := self.serviceRpc(&request, requestType, &response, responseType)
	if err != nil {
		return nil, err
	}

	if response.GetStatus() == consensus_pb2.ConsensusStateGetResponse_UNKNOWN_BLOCK {
		return nil, &UnknownBlockError{
			Msg: "Block not found",
		}
	} else if response.GetStatus() != consensus_pb2.ConsensusStateGetResponse_OK {
		return nil, &ReceiveError{
			Msg: fmt.Sprintf("Failed with status: %v", response.GetStatus()),
		}
	}

	result := make(map[string][]byte, len(response.GetEntries()))
	for _, entry := range response.GetEntries() {
		result[entry.GetAddress()] = entry.GetData()
	}
	return result, nil
}

// serviceRpc is a general purpose RPC routine to implement service requests from the consensus engine, back to the validator.
func (self *ZmqService) serviceRpc(request proto.Message, requestType validator_pb2.Message_MessageType, response proto.Message, responseType validator_pb2.Message_MessageType) error {
	// Marshal the request message
	bytes, err := proto.Marshal(request)
	if err != nil {
		return fmt.Errorf("Failed to marshal %v message: %v", requestType, err)
	}

	// Send the message
	corrId, err := self.connection.SendNewMsg(requestType, bytes)
	if err != nil {
		return fmt.Errorf("Failed to send %v message: %v", requestType, err)
	}
	self.corrIds[corrId] = nil

	// Receive the response
	// TODO: Do we need to do a timeout here?
	_, msg, err := self.connection.RecvMsgWithId(corrId)
	if msg.GetCorrelationId() != corrId {
		return fmt.Errorf("Expected message with correlation id %v but received %v", corrId, msg.GetCorrelationId())
	}
	delete(self.corrIds, corrId)

	// Check the response type
	if msg.GetMessageType() != responseType {
		return fmt.Errorf("Expected message with type %v but received %v", responseType, msg.GetMessageType())
	}

	// Parse the response message
	err = proto.Unmarshal(msg.GetContent(), response)
	if err != nil {
		return fmt.Errorf("Failed to unmarshal %v message: %v", responseType, err)
	}

	return nil
}

// hasCorrelationId returns whether this ZmqService contains a given CorrelationId
func (self *ZmqService) hasCorrelationId(corrId string) bool {
	_, exists := self.corrIds[corrId]
	return exists
}
