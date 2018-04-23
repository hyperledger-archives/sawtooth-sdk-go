*Build go sdk*
```
go get github.com/hyperledger/sawtooth-sdk-go
cd $GOPATH/src/github.com/hyperledger/sawtooth-sdk-go
go generate
```
Docker instructions
```
cd sawtooth-sdk-go
docker build . -t sawtooth-sdk-go
docker run -v $(pwd):/project/sawtooth-sdk-go sawtooth-sdk-go
```

Go generate will build the protos / mocks and place them in the protobuf or mocks directory respectively.
