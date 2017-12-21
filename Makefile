OLDGOPATH := $(GOPATH)/src/
GOPATH := $(GOPATH):$(PWD)
-include config.mk

all: bin bin/rcoind$(EXE)
bin:
	mkdir -p bin
bin/rcoind$(EXE): bin $(wildcard src/rcoin/*.go)
	go build -i -o bin/rcoind rcoin

deps:
	go get -d golang.org/x/crypto/ed25519
	go get -d github.com/syndtr/goleveldb/leveldb
	go get -d golang.org/x/crypto/scrypt
	go get -d github.com/vmihailenco/msgpack
	go get -d github.com/ccding/go-stun/stun
	go get -d github.com/NebulousLabs/go-upnp

dist-binaries:
	mkdir -p dist
	env GOOS=windows GOARCH=386 go build -i -o dist/rcoind.exe rcoin
	env GOOS=linux GOARCH=386 go build -i -o dist/rcoind-linux386 rcoin
	env GOOS=darwin GOARCH=amd64 go build -i -o dist/rcoind-macosx rcoin
