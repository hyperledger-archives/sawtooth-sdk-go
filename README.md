*Build go sdk*
```
go get github.com/hyperledger/sawtooth-sdk-go
cd $GOPATH/src/github.com/hyperledger/sawtooth-sdk-go
```
Docker instructions
```
cd sawtooth-sdk-go
docker build . -t sawtooth-sdk-go
docker run -v $(pwd):/go/src/github.com/hyperledger/sawtooth-sdk-go sawtooth-sdk-go
```

Updating Protocol Buffers

When maintainers and contributors wish to update the protobuf definitions, they
will need to run `go generate` after making any changes. `go generate` will
remove the pre-existing protobuf directory, and generate a new protobuf directory
based on `.proto` files in the protos directory.
