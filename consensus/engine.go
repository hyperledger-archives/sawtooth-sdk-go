package consensus

import (
	"fmt"
	"os"
	"os/signal"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/hyperledger/sawtooth-sdk-go/logging"
	"github.com/hyperledger/sawtooth-sdk-go/messaging"
	"github.com/hyperledger/sawtooth-sdk-go/protobuf/consensus_pb2"
	"github.com/hyperledger/sawtooth-sdk-go/protobuf/network_pb2"
	"github.com/hyperledger/sawtooth-sdk-go/protobuf/validator_pb2"
	zmq "github.com/pebbe/zmq4"
)

var logger *logging.Logger = logging.Get()

const REGISTER_TIMEOUT = 300
const INITIAL_RETRY_DELAY = time.Millisecond * 100
const MAX_RETRY_DELAY = time.Second * 3

// A ConsensusEngine implements the bulk of the behavior required to interface a pluggable consensus engine
// with the Sawtooth validator.
type ConsensusEngine struct {
	uri        string
	impl       ConsensusEngineImpl
	shutdownTx messaging.Connection

	notifyChan chan Notification

	service *ZmqService
}

// NewConsensusEngine returns a new ConsensusEngine instance, given the URI of a Sawtooth validator consensus
// port and an implementation of ConsensusEngineImpl.
func NewConsensusEngine(uri string, impl ConsensusEngineImpl) *ConsensusEngine {
	return &ConsensusEngine{
		uri:  uri,
		impl: impl,
	}
}

// Start is called after the ConsensusEngine has been created, and begins communcation with the validator. Start
// does not return until the consensus engine terminates.
func (self *ConsensusEngine) Start() error {
	for {
		context, err := zmq.NewContext()
		if err != nil {
			panic(fmt.Sprintf("Failed to create ZMQ context: %s", err))
		}

		reconnect, err := self.start(context)
		if err != nil {
			return err
		}

		if !reconnect {
			break
		}
	}
	return nil
}

// start continuously handles messages that arrive to the conesnsus engine.
func (self *ConsensusEngine) start(context *zmq.Context) (bool, error) {
	var err error

	restart := false

	// Main connection to the validator
	// All messages, including responses to service requests come via this connection
	validator, err := messaging.NewConnection(context, zmq.DEALER, self.uri, false)
	if err != nil {
		return restart, fmt.Errorf("Could not connect to validator: %v", err)
	}

	// Monitor to determine if the validator disconnects
	monitor, err := validator.Monitor(zmq.EVENT_DISCONNECTED)
	if err != nil {
		return restart, fmt.Errorf("Could not monitor validator connection: %v", err)
	}

	// Connection to receive requests from the consensus service
	service, err := messaging.NewConnection(context, zmq.PAIR, "inproc://service", true)
	if err != nil {
		return restart, fmt.Errorf("Could not create service router: %v", err)
	}

	// Internal connection to manage shutdown.
	shutdownRx, err := messaging.NewConnection(context, zmq.PAIR, "inproc://shutdown", true)
	if err != nil {
		return restart, fmt.Errorf("Could not create shutdown connection: %v", err)
	}
	self.shutdownTx, err = messaging.NewConnection(context, zmq.PAIR, "inproc://shutdown", false)

	// Register engine with validator
	logger.Infof("Registering (%v, %v)", self.impl.Name(), self.impl.Version())

	// Variable to hold startup start received from the validator
	var startupState StartupState

	// We need to check if register() returns startup state, otherwise call waitUntilActive()
	// Different versions of Sawtooth behave differently here
	startupState, err = self.register(validator, shutdownRx)
	if err != nil {
		return restart, fmt.Errorf("Error registering engine (%v, %v): %v", self.impl.Name(), self.impl.Version(), err)
	}

	// If startupState is nil, we didn't receive it from register(), so call waitUntilActive()
	if startupState == nil {
		startupState, err = self.waitUntilActive(validator, shutdownRx)
		if err != nil {
			return restart, fmt.Errorf("Error waiting for engine activation (%v, %v): %v", self.impl.Name(), self.impl.Version(), err)
		}
	}

	logger.Info("Consensus engine is active")

	// Initialize a channel to pass notifications to the engine impl
	self.notifyChan = make(chan Notification, 2)

	// Create and start the service
	self.service = NewZmqService(context, "inproc://service")
	self.service.Start()

	// Start the impl in its own goroutine.
	go self.impl.Start(startupState, self.service, self.notifyChan)

	// Set up ZMQ polling across the various components
	poller := zmq.NewPoller()
	poller.Add(validator.Socket(), zmq.POLLIN)
	poller.Add(monitor, zmq.POLLIN)
	poller.Add(service.Socket(), zmq.POLLIN)
	poller.Add(shutdownRx.Socket(), zmq.POLLIN)

	// Main polling loop
	// This selects between components that have messages or notifications
	// ready to receive and services those events accordingly
	logger.Debug("Starting main polling loop")
	for {
		polled, err := poller.Poll(-1)
		if err != nil {
			return restart, fmt.Errorf("Polling failed: %v", err)
		}

		for _, ready := range polled {
			switch socket := ready.Socket; socket {
			case validator.Socket():
				self.receiveValidator(validator, service)
			case service.Socket():
				self.receiveService(validator, service)
			case monitor:
				restart = self.receiveMonitor(monitor, self.shutdownTx)
			case shutdownRx.Socket():
				restart = self.receiveShutdown(shutdownRx, validator, service, monitor)
				return restart, nil
			}
		}
	}
}

// Shutdown stops the ConsensusEngine.
func (self *ConsensusEngine) Shutdown() {
	// Initiate a clean shutdown
	if self.shutdownTx != nil {
		err := self.shutdownTx.SendData("", []byte("shutdown"))
		if err != nil {
			logger.Errorf("Failed to send shutdown command")
		}
	}
}

// ShutdownOnSignal configures the ConsensusEngine to shutdown upon receiving a certain signal.
func (self *ConsensusEngine) ShutdownOnSignal(siglist ...os.Signal) {
	// Setup signal handlers
	ch := make(chan os.Signal)
	signal.Notify(ch, siglist...)

	go func() {
		// Wait for a signal
		_ = <-ch

		// Reset signal handlers
		signal.Reset(siglist...)
		logger.Warnf("Shutting down gracefully (Press Ctrl+C again to force)")
		self.Shutdown()
	}()
}

// register performs the initial registration of the consensus engine with
// the validator. Returns StartupState which contains information about the
// validator's current state.
func (self *ConsensusEngine) register(validator, shutdownRx messaging.Connection) (StartupState, error) {
	regRequest := &consensus_pb2.ConsensusRegisterRequest{
		Name:    self.impl.Name(),
		Version: self.impl.Version(),
	}

	regRequestData, err := proto.Marshal(regRequest)
	if err != nil {
		return nil, err
	}

	_, err = validator.SendNewMsg(
		validator_pb2.Message_CONSENSUS_REGISTER_REQUEST,
		regRequestData,
	)
	if err != nil {
		return nil, err
	}

	poller := zmq.NewPoller()
	poller.Add(validator.Socket(), zmq.POLLIN)
	poller.Add(shutdownRx.Socket(), zmq.POLLIN)

	logger.Debug("register: about to poll")
	for {
		polled, err := poller.Poll(REGISTER_TIMEOUT)
		if err != nil {
			return nil, fmt.Errorf("Polling failed: %v", err)
		}

		retry_delay := INITIAL_RETRY_DELAY
		for _, ready := range polled {
			switch socket := ready.Socket; socket {
			case validator.Socket():
				var msg *validator_pb2.Message
				_, msg, err := validator.RecvMsg()
				if err != nil {
					return nil, err
				}

				if msg.GetMessageType() != validator_pb2.Message_CONSENSUS_REGISTER_RESPONSE {
					return nil, &ReceiveError{Msg: fmt.Sprintf("Received unexpected message type: %v", msg.GetMessageType())}
				}

				regResponse := &consensus_pb2.ConsensusRegisterResponse{}
				err = proto.Unmarshal(msg.GetContent(), regResponse)
				if err != nil {
					return nil, err
				}

				switch regResponse.GetStatus() {
				case consensus_pb2.ConsensusRegisterResponse_OK:
					// Handle successful registration
					logger.Infof(
						"Successfully registered engine (%v, %v)",
						self.impl.Name(),
						self.impl.Version(),
					)

					if regResponse.GetChainHead() != nil && regResponse.GetLocalPeerInfo() != nil {
						startupState := newStartupStateFromProtos(regResponse.GetChainHead(), regResponse.GetPeers(), regResponse.GetLocalPeerInfo())
						return startupState, nil
					} else {
						return nil, nil
					}

				case consensus_pb2.ConsensusRegisterResponse_NOT_READY:
					// Handle not ready
					logger.Infof(
						"Validator returned not ready on registration for engine (%v, %v)",
						self.impl.Name(),
						self.impl.Version(),
					)

					time.Sleep(retry_delay)
					if retry_delay < MAX_RETRY_DELAY {
						retry_delay *= 2
						if retry_delay > MAX_RETRY_DELAY {
							retry_delay = MAX_RETRY_DELAY
						}
					}

					_, err = validator.SendNewMsg(
						validator_pb2.Message_CONSENSUS_REGISTER_REQUEST,
						regRequestData,
					)

					continue
				default:
					return nil, &ReceiveError{Msg: fmt.Sprintf("Registration failed with status: %v", regResponse.GetStatus())}
				}

			case shutdownRx.Socket():
				return nil, fmt.Errorf("Received shutdown while registering")
			}
		}
	}
}

// waitUntilActive waits for the validator to send an activation message containing
// startup state. This method is only applicable with validator versions > v1.1.
func (self *ConsensusEngine) waitUntilActive(validator, shutdownRx messaging.Connection) (StartupState, error) {
	poller := zmq.NewPoller()
	poller.Add(validator.Socket(), zmq.POLLIN)
	poller.Add(shutdownRx.Socket(), zmq.POLLIN)

	logger.Debug("waitUntilActive: about to poll")
	for {
		polled, err := poller.Poll(time.Millisecond * 100)
		if err != nil {
			return nil, fmt.Errorf("Polling failed: %v", err)
		}

		for _, ready := range polled {
			switch socket := ready.Socket; socket {
			case validator.Socket():
				var msg *validator_pb2.Message
				_, msg, err := validator.RecvMsg()
				if err != nil {
					return nil, err
				}

				if msg.GetMessageType() != validator_pb2.Message_CONSENSUS_NOTIFY_ENGINE_ACTIVATED {
					return nil, &ReceiveError{Msg: fmt.Sprintf("Received unexpected message type: %v", msg.GetMessageType())}
				}

				activationResponse := &consensus_pb2.ConsensusNotifyEngineActivated{}
				err = proto.Unmarshal(msg.GetContent(), activationResponse)
				if err != nil {
					return nil, err
				}

				regAck := &consensus_pb2.ConsensusNotifyAck{}
				regAckBytes, err := proto.Marshal(regAck)
				if err != nil {
					return nil, err
				}

				err = validator.SendMsg(validator_pb2.Message_CONSENSUS_NOTIFY_ACK, regAckBytes, msg.GetCorrelationId())
				if err != nil {
					return nil, err
				}

				logger.Infof(
					"Successfully activated engine (%v, %v)",
					self.impl.Name(),
					self.impl.Version(),
				)

				startupState := newStartupStateFromProtos(activationResponse.GetChainHead(), activationResponse.GetPeers(), activationResponse.GetLocalPeerInfo())

				return startupState, nil

			case shutdownRx.Socket():
				return nil, fmt.Errorf("Received shutdown while waiting for activation")
			}
		}
	}
}

// receiveValidator handles messages from the validator.
func (self *ConsensusEngine) receiveValidator(validator, service messaging.Connection) {
	// Receive a message from the validator
	_, data, err := validator.RecvData()
	if err != nil {
		logger.Errorf("Receiving message from validator failed: %v", err)
		return
	}
	// Deserialize to get the correlation id
	msg, err := messaging.LoadMsg(data)
	if err != nil {
		logger.Errorf("Deserializing message from validator failed: %v", err)
		return
	}

	// Get the message type
	t := msg.GetMessageType()
	// Get the correlation id
	corrId := msg.GetCorrelationId()

	// Switch on the message type
	switch t {
	case
		validator_pb2.Message_CONSENSUS_NOTIFY_PEER_CONNECTED,
		validator_pb2.Message_CONSENSUS_NOTIFY_PEER_DISCONNECTED,
		validator_pb2.Message_CONSENSUS_NOTIFY_PEER_MESSAGE,
		validator_pb2.Message_CONSENSUS_NOTIFY_BLOCK_NEW,
		validator_pb2.Message_CONSENSUS_NOTIFY_BLOCK_VALID,
		validator_pb2.Message_CONSENSUS_NOTIFY_BLOCK_INVALID,
		validator_pb2.Message_CONSENSUS_NOTIFY_BLOCK_COMMIT:

		// Process the notification and pass it to the impl
		self.handleNotification(msg)

		// Send an ACK to the validator
		data, err := proto.Marshal(&consensus_pb2.ConsensusNotifyAck{})
		if err != nil {
			logger.Errorf("Failed to marshal consensus update ack: %v", err)
			return
		}
		err = validator.SendMsg(validator_pb2.Message_CONSENSUS_NOTIFY_ACK, data, corrId)
		if err != nil {
			logger.Errorf("Failed to send consensus notify ack: %v", err)
			return
		}

		return

	case validator_pb2.Message_PING_REQUEST:
		data, err := proto.Marshal(&network_pb2.PingResponse{})
		if err != nil {
			logger.Errorf("Failed to marshal ping response: %v", err)
			return
		}
		err = validator.SendMsg(validator_pb2.Message_PING_RESPONSE, data, corrId)
		if err != nil {
			logger.Errorf("Failed to send ping response: %v", err)
			return
		}

		logger.Info("Responded to ping")
		return

	// Any other message type is most likely a response to something sent by the service
	default:
		if corrId != "" && self.service.hasCorrelationId(corrId) {
			err = service.SendData("", data)
			if err != nil {
				logger.Errorf("Failed to send service response with correlation id %v: %v", corrId, err)
				return
			}
			return
		} else {
			logger.Warnf("Received unexpected message from validator: (%v, %v)", t, corrId)
		}
	}
}

// receiveService handles messages from the service.
func (self *ConsensusEngine) receiveService(validator, service messaging.Connection) {
	// Receive a message from the service
	_, data, err := service.RecvData()
	if err != nil {
		logger.Errorf("Receiving message from service failed: %v", err)
		return
	}

	msg, err := messaging.LoadMsg(data)
	if err != nil {
		logger.Errorf("Deserializing message from service failed: %v", err)
		return
	}

	// Pass the message on to the validator
	err = validator.SendData("", data)
	if err != nil {
		logger.Errorf("Failed to send message (%v) to validator: %v", msg.GetCorrelationId(), err)
		return
	}
}

// handleNotification handles an incoming protobuf message from the validator.
func (self *ConsensusEngine) handleNotification(msg *validator_pb2.Message) {
	unmarshalHelper := func(msg *validator_pb2.Message, message proto.Message) error {
		err := proto.Unmarshal(msg.GetContent(), message)
		if err != nil {
			return fmt.Errorf("Failed to unmarshal %v: %v", msg.GetMessageType(), err)
		}
		return nil
	}

	logger.Debug(msg.GetMessageType())

	switch msg.GetMessageType() {
	case validator_pb2.Message_CONSENSUS_NOTIFY_PEER_CONNECTED:
		notificationProto := consensus_pb2.ConsensusNotifyPeerConnected{}
		err := unmarshalHelper(msg, &notificationProto)
		if err != nil {
			logger.Error(err)
			break
		}

		notification := NotificationPeerConnected{
			PeerInfo: &peerInfo{
				peerId: NewPeerIdFromSlice(notificationProto.GetPeerInfo().PeerId),
			},
		}

		self.notifyChan <- notification

	case validator_pb2.Message_CONSENSUS_NOTIFY_PEER_DISCONNECTED:
		notificationProto := consensus_pb2.ConsensusNotifyPeerDisconnected{}
		err := unmarshalHelper(msg, &notificationProto)
		if err != nil {
			logger.Error(err)
			break
		}

		notification := NotificationPeerDisconnected{
			PeerInfo: &peerInfo{
				peerId: NewPeerIdFromSlice(notificationProto.GetPeerId()),
			},
		}

		self.notifyChan <- notification

	case validator_pb2.Message_CONSENSUS_NOTIFY_PEER_MESSAGE:
		notificationProto := consensus_pb2.ConsensusNotifyPeerMessage{}
		err := unmarshalHelper(msg, &notificationProto)
		if err != nil {
			logger.Error(err)
			break
		}

		notification := NotificationPeerMessage{
			PeerMessage: newPeerMessageFromProto(notificationProto.GetMessage()),
			SenderId:    NewPeerIdFromSlice(notificationProto.GetSenderId()),
		}

		self.notifyChan <- notification

	case validator_pb2.Message_CONSENSUS_NOTIFY_BLOCK_NEW:
		notificationProto := consensus_pb2.ConsensusNotifyBlockNew{}
		err := unmarshalHelper(msg, &notificationProto)
		if err != nil {
			logger.Error(err)
			break
		}

		notification := NotificationBlockNew{
			Block: newBlockFromProto(notificationProto.GetBlock()),
		}

		self.notifyChan <- notification

	case validator_pb2.Message_CONSENSUS_NOTIFY_BLOCK_VALID:
		notificationProto := consensus_pb2.ConsensusNotifyBlockValid{}
		err := unmarshalHelper(msg, &notificationProto)
		if err != nil {
			logger.Error(err)
			break
		}

		notification := NotificationBlockValid{
			BlockId: NewBlockIdFromBytes(notificationProto.GetBlockId()),
		}

		self.notifyChan <- notification

	case validator_pb2.Message_CONSENSUS_NOTIFY_BLOCK_INVALID:
		notificationProto := consensus_pb2.ConsensusNotifyBlockInvalid{}
		err := unmarshalHelper(msg, &notificationProto)
		if err != nil {
			logger.Error(err)
			break
		}

		notification := NotificationBlockInvalid{
			BlockId: NewBlockIdFromBytes(notificationProto.GetBlockId()),
		}

		self.notifyChan <- notification

	case validator_pb2.Message_CONSENSUS_NOTIFY_BLOCK_COMMIT:
		notificationProto := consensus_pb2.ConsensusNotifyBlockCommit{}
		err := unmarshalHelper(msg, &notificationProto)
		if err != nil {
			logger.Error(err)
			break
		}

		notification := NotificationBlockCommit{
			BlockId: NewBlockIdFromBytes(notificationProto.GetBlockId()),
		}

		self.notifyChan <- notification
	}
}

// receiveMonitor handles monitor events.
func (self *ConsensusEngine) receiveMonitor(monitor *zmq.Socket, shutdownTx messaging.Connection) bool {
	restart := false
	event, endpoint, _, err := monitor.RecvEvent(0)
	if err != nil {
		logger.Error(err)
	} else {
		if event == zmq.EVENT_DISCONNECTED {
			logger.Infof("Validator '%v' disconnected", endpoint)
			restart = true
		} else {
			logger.Errorf("Received unexpected event on monitor socket: %v", event)
		}
		err = shutdownTx.SendData("", []byte("restart"))
		if err != nil {
			logger.Errorf("Failed to send restart command")
		}
	}
	return restart
}

// receiveShutdown handles shutdown commands.
func (self *ConsensusEngine) receiveShutdown(shutdownRx, validator, service messaging.Connection, monitor *zmq.Socket) bool {
	_, data, err := shutdownRx.RecvData()
	if err != nil {
		logger.Errorf("Error receiving shutdown: %v", err)
	}

	restart := false
	cmd := string(data)
	if cmd == "restart" {
		restart = true
	} else if cmd != "shutdown" {
		fmt.Errorf("Received unexpected shutdown command: %v", cmd)
		return false
	}

	logger.Infof("Received command to %v", cmd)
	err = self.performShutdown(shutdownRx, validator, service, monitor)
	if err != nil {
		logger.Errorf("Error shutting down: %v", err)
		return false
	}

	return restart
}

// performShutdown shuts down the ConsensusEngine.
func (self *ConsensusEngine) performShutdown(shutdownRx, validator, service messaging.Connection, monitor *zmq.Socket) error {
	self.notifyChan <- &NotificationShutdown{}
	close(self.notifyChan)
	self.service.Shutdown()

	monitor.Close()
	validator.Close()
	service.Close()
	shutdownRx.Close()

	return nil
}
