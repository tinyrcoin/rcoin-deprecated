GOPATH2 := $(shell echo $$GOPATH)
export GOPATH = $(GOPATH2):$(shell pwd)
-include config.mk
#
CONFIRM = n
all: bin bin/rcoind$(EXE)
debug-gopath:
	echo Original gopath: $(GOPATH2)
	echo $(GOPATH)
bin:
	mkdir -p bin
bin/rcoind$(EXE): bin $(wildcard src/rcoin/*.go)
	go build -i -o bin/rcoind rcoin

deps:
	go get -d github.com/satori/go.uuid
	go get -d golang.org/x/crypto/ed25519
	go get -d github.com/syndtr/goleveldb/leveldb
	go get -d golang.org/x/crypto/scrypt
	go get -d github.com/vmihailenco/msgpack
	go get -d github.com/ccding/go-stun/stun
	go get -d github.com/NebulousLabs/go-upnp
reset:
	echo "Are you sure? CONFIRM=$(CONFIRM)"
	/usr/bin/test "$(CONFIRM)" == "y" || false
	rm -r $(HOME)/.rcoin/rcoin.db
	rm $(HOME)/.rcoin/peers.txt
dist-binaries-dir:
	mkdir -p dist/win32 dist/linux dist/mac dist/freebsd
dist-binaries: dist-binaries-dir dist-win32 dist-linux dist-mac
dist-win32:
	env GOOS=windows GOARCH=386 go build -i -o dist/win32/rcoind.exe rcoin
dist-linux:
	env CGO_ENABLED=0 GOOS=linux GOARCH=386 go build -i -o dist/linux/rcoind rcoin
dist-mac:
	env CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -i -o dist/mac/rcoind rcoin
dist-freebsd:
	env CGO_ENABLED=0 GOOS=freebsd GOARCH=amd64 go build -i -o dist/freebsd/rcoind rcoin
