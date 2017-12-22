GOPATH2 := $(shell echo $$GOPATH)
GOPATH := $(GOPATH2):$(shell pwd)
export GOPATH
-include config.mk
#
all: bin bin/rcoind$(EXE)
debug-gopath:
	echo Original gopath: $(GOPATH2)
	echo $(GOPATH)
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

dist-binaries-dir:
	mkdir -p dist
dist-binaries: dist-binaries-dir dist-win32 dist-linux dist-mac
dist-win32:
	env GOOS=windows GOARCH=386 go build -i -o dist/rcoind.exe rcoin
dist-linux:
	env CGO_ENABLED=0 GOOS=linux GOARCH=386 go build -i -o dist/rcoind-linux386 rcoin
dist-mac:
	env CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -i -o dist/rcoind-macosx rcoin
