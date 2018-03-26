*Build go sdk*
```
go get github.com/hyperledger/sawtooth-sdk-go
cd $GOPATH/src/github.com/hyperledger/sawtooth-sdk-go
go generate
```
Go generate will preform a sparse checkout and build the the protos from sawtooth-core.
