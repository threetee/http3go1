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
	go test admin.go
	go test redirector.go

clean:
	rm -f $(ALL)
	rm -rf $(CLEANFILES)
	go clean

admin:
	go build cmd/admin/admin.go

redirector:
	go build cmd/redirector/redirector.go

docker-dist: docker-dist-redirector docker-dist-admin

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

docker-push:
	docker push threetee/http3go1-redirector
	docker push threetee/http3go1-admin

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
