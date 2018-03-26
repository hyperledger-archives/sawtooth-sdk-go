//go:generate /bin/mkdir -p protobuf dependencies/sawtooth-core
//go:generate /bin/bash -c "cd dependencies/sawtooth-core; git init; git remote add -f origin https://github.com/hyperledger/sawtooth-core.git"
//go:generate /bin/bash -c "cd dependencies/sawtooth-core; git config core.sparseCheckout true; echo 'protos' > .git/info/sparse-checkout; git pull --depth=1 origin master"
//go:generate /bin/bash -c "/usr/bin/find dependencies/sawtooth-core/protos/ -type f -exec protoc -I=dependencies/sawtooth-core/protos --go_out=protobuf {} \\;"

package gen
