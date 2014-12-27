ALL=admin redirector

all: $(ALL)

%: %.go
	go build $@.go

clean:
	rm -f $(ALL)
	go clean
