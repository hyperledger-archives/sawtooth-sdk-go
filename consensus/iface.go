package consensus

type ConsensusService interface {
	SendTo(peer PeerId, messageType string, payload []byte) error
	Broadcast(messageType string, payload []byte) error

	InitializeBlock(previousId BlockId) error
	SummarizeBlock() ([]byte, error)
	FinalizeBlock(data []byte) (BlockId, error)
	CancelBlock() error

	CheckBlocks(blockIds []BlockId) error
	CommitBlock(blockId BlockId) error
	IgnoreBlock(blockId BlockId) error
	FailBlock(blockId BlockId) error

	GetBlocks(blockIds []BlockId) (map[BlockId]Block, error)
	GetChainHead() (Block, error)
	GetSettings(blockId BlockId, keys []string) (map[string]string, error)
	GetState(blockId BlockId, addresses []string) (map[string][]byte, error)
}

type ConsensusEngineImpl interface {
	Name() string
	Version() string

	Start(startupState StartupState, service ConsensusService, notifyChan chan Notification) error
}
