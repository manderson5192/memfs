help:
	@echo "Welcome to MemFS!  Here's a list of available Makefile targets:"
	@echo ""
	@$(MAKE) list-targets

run:
	go run main.go

clean:
	rm -rf build

build:
	go build -o build/main main.go

test:
	go test -v ./...

# Lists all available targets within the Makefile, per https://stackoverflow.com/a/26339924
.PHONY: list-targets
list-targets:
	@$(MAKE) -pRrq -f $(lastword $(MAKEFILE_LIST)) : 2>/dev/null | awk -v RS= -F: '/^# File/,/^# Finished Make data base/ {if ($$1 !~ "^[#.]") {print $$1}}' | sort | egrep -v -e '^[^[:alnum:]]' -e '^$@$$'