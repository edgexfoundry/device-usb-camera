GO=go
HOOKS=cmd/configure/configure cmd/install/install

tidy:
	go mod tidy -compat=1.17

build: $(HOOKS)

cmd/configure/configure:
	$(GO) build -o $@ ./cmd/configure

cmd/install/install:
	$(GO) build -o $@ ./cmd/install

clean:
	rm -f $(HOOKS)

