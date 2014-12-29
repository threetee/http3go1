ifeq ($(MYTARGDIR),)
	MYTARGDIR:=target
endif

ifeq ($(PREFIX),)
	PREFIX:=usr
endif

ifeq ($(STATIC_DIR),)
	STATIC_DIR:=$(PREFIX)/share/http3go1
endif

CLEANFILES=$(MYTARGDIR)

ALL=admin redirector

all: $(ALL)

%: %.go
	go build $@.go

test:
	go test lib/*.go
	go test admin.go
	go test redirector.go

clean:
	rm -f $(ALL)
	rm -rf $(CLEANFILES)
	go clean

bin-dist: admin redirector assets
	cp admin $(MYTARGDIR)/$(PREFIX)/bin
	cp redirector $(MYTARGDIR)/$(PREFIX)/bin
	git log --pretty=format:"http3go1 %H" -1 > $(MYTARGDIR)/$(STATIC_DIR)/_version

assets: directories
	cp -r static/* $(MYTARGDIR)/$(STATIC_DIR)

directories:
	mkdir -p $(MYTARGDIR)/$(STATIC_DIR)
	mkdir -p $(MYTARGDIR)/$(PREFIX)/bin

.PHONY: all clean assets directories test
