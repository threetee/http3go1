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

ALL=deps admin redirector

all: $(ALL)

deps:
	go get github.com/kelseyhightower/envconfig
	go get github.com/simonz05/godis/redis
	go get github.com/gorilla/mux
	go get github.com/golang/glog

%: deps %.go
	go build $@.go

test: deps
	go test lib/*.go
	go test admin.go
	go test redirector.go

clean:
	rm -f $(ALL)
	rm -rf $(CLEANFILES)
	go clean

docker-dist: deps docker-dist-redirector docker-dist-admin

docker-dist-redirector:
	-rm -f Dockerfile
	ln -s Dockerfile.redirector Dockerfile
	docker build -t threetee/http3go1-redirector .
	-rm -f Dockerfile

docker-dist-admin:
	-rm -f Dockerfile
	ln -s Dockerfile.admin Dockerfile
	docker build -t threetee/http3go1-admin .
	-rm -f Dockerfile

bin-dist: admin redirector assets
	cp admin $(MYTARGDIR)/$(PREFIX)/bin
	cp redirector $(MYTARGDIR)/$(PREFIX)/bin
	git log --pretty=format:"http3go1 %H" -1 > $(MYTARGDIR)/$(STATIC_DIR)/_version

assets: directories
	cp -r static/* $(MYTARGDIR)/$(STATIC_DIR)

directories:
	mkdir -p $(MYTARGDIR)/$(STATIC_DIR)
	mkdir -p $(MYTARGDIR)/$(PREFIX)/bin

.PHONY: all deps clean assets directories test
