module github.com/hyperledger/sawtooth-devmode-go/examples/devmode_consensus

go 1.16

replace github.com/hyperledger/sawtooth-sdk-go => ../../

require (
	github.com/hyperledger/sawtooth-sdk-go v0.0.0-00010101000000-000000000000
	github.com/jessevdk/go-flags v1.4.0
)
