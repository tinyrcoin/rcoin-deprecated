SDX ?= sdx.kit
PLAT ?= win32
PLATPATH ?= ../dist
KITS ?= tclkits/$(PLAT)
tkwallet$(PLAT): wallet.tcl starkit_pre.tcl
	rm -frrr dist/tkwallet$(PLAT)
	mkdir -p dist/tkwallet$(PLAT)
	cp $(KITS)/* dist/tkwallet$(PLAT) 
	cat starkit_pre.tcl wallet.tcl > /tmp/tkw001.tcl
	$(SDX) qwrap /tmp/tkw001.tcl -runtime dist/tkwallet$(PLAT)/tkwallet*
	mv tkw001 dist/tkwallet$(PLAT)/tkwallet*
	cp $(PLATPATH)/$(PLAT)/* dist/tkwallet$(PLAT)
	if [ "$(PLAT)" = "win32" ]; then cd dist/tkwallet$(PLAT); rm ../tkwallet.zip; zip -9r ../tkwallet.zip .; fi

debug-plat:
	echo $(shell echo $$PLATwin)
