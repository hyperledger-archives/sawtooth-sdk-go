![Hyperledger Sawtooth](https://raw.githubusercontent.com/hyperledger/sawtooth-core/master/images/sawtooth_logo_light_blue-small.png) 
# Hyperledger Sawtooth SDK for Go

*Hyperledger Sawtooth* is an enterprise solution for building, deploying, and
running distributed ledgers (also called blockchains). It provides an
extremely modular and flexible platform for implementing transaction-based
updates to shared state between untrusted parties coordinated by consensus
algorithms.

The *Sawtooth SDK for Go* provides a number of useful components that simplify
developing Go applications which interface with the Sawtooth platform.
These include modules to sign Transactions, read state, and create Transaction
Processors. For full usage and installation instructions please reference the
official Sawtooth documentation below:

* [Hyperledger Sawtooth Official Documentation](https://sawtooth.hyperledger.org/docs/)
* [Sawtooth Application Developer's Guide](https://sawtooth.hyperledger.org/docs/core/releases/latest/app_developers_guide.html)
* [Sawtooth Go SDK Guide](https://sawtooth.hyperledger.org/docs/core/releases/latest/app_developers_guide/go_sdk.html)

## Quickstart Instructions

To start using the SDK in your project, simply run inside your project root:
```
go get github.com/hyperledger/sawtooth-sdk-go
```
As of `v0.1.4`, this SDK is a *Go Module* and can simply be installed and without a further code generation step.
For earlier versions, you must follow the instructions below to generate `protobuf` code with `go generate`.

## Developer Instructions

### Updating Dependencies

If any update requires any additional or upgraded dependencies, you should do the upgrade using `go get`, then run
`go mod tidy`. In the case that any dependency added or modified, make sure to commit the updated `go.mod` and `go.sum`
files.

### Updating Protocol Buffers

When maintainers and contributors wish to update the `protobuf` definitions, they
will need to run `go generate` after making any changes. `go generate` will
remove the pre-existing protobuf directory, and generate a new protobuf directory
based on `.proto` files in the protos directory. Any files modified in the `protobuf`
directory should be committed.

### Running `go generate`

To run `go generate`, you should simply change to the repo directory and execute the command:
```
cd sawtooth-sdk-go
go generate
```
The `generate` scripts do have some dependencies that can be non-trivial to install on some systems. In this
case, you can use Docker to perform the generation as follows:
```
cd sawtooth-sdk-go
docker build . -t sawtooth-sdk-go
docker run -v $(pwd):/go/src/github.com/hyperledger/sawtooth-sdk-go sawtooth-sdk-go
```
