PROTOBUFF_DIR=${PWD}/pkg/generated

PROTOBUFF_SDK_DIR=${PROTOBUFF_DIR}/sawtooth/sdk/protobuf
PROTO_SDK_DIR=${PWD}/protos/sawtooth/sdk/protobuf

PROTOBUFF_BLOCK_INFO_DIR=${PROTOBUFF_DIR}/sawtooth/blockinfo/protobuf
PROTO_BLOCK_INFO_DIR=${PWD}/protos/sawtooth/blockinfo/protobuf

PROTOBUFF_IDENTITY_DIR=${PROTOBUFF_DIR}/sawtooth/identity/protobuf
PROTO_IDENTITY_DIR=${PWD}/protos/sawtooth/identity/protobuf

PROTOBUFF_SMALL_BANK_DIR=${PROTOBUFF_DIR}/sawtooth/smallbank/protobuf
PROTO_SMALL_BANK_DIR=${PWD}/protos/sawtooth/smallbank/protobuf

.PHONY: all sdk block_info identity smallbank
all: sdk block_info identity smallbank

sdk:
	mkdir -p ${PROTOBUFF_SDK_DIR}
	protoc --proto_path=${PROTO_SDK_DIR} --go_out=${PROTOBUFF_SDK_DIR} ${PROTO_SDK_DIR}/*.proto

block_info:
	mkdir -p ${PROTOBUFF_BLOCK_INFO_DIR}
	protoc --proto_path=${PROTO_BLOCK_INFO_DIR} --go_out=${PROTOBUFF_BLOCK_INFO_DIR} ${PROTO_BLOCK_INFO_DIR}/*.proto

identity:
	mkdir -p ${PROTOBUFF_IDENTITY_DIR}
	protoc --proto_path=${PROTO_IDENTITY_DIR} --go_out=${PROTOBUFF_IDENTITY_DIR} ${PROTO_IDENTITY_DIR}/*.proto

smallbank:
	mkdir -p ${PROTOBUFF_SMALL_BANK_DIR}
	protoc --proto_path=${PROTO_SMALL_BANK_DIR} --go_out=${PROTOBUFF_SMALL_BANK_DIR} ${PROTO_SMALL_BANK_DIR}/*.proto


clean:
	rm -rf ${PROTOBUFF_DIR}
