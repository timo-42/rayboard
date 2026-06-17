APP := rayboard
CMD := ./cmd/rayboard
DIST := dist

.PHONY: test
test:
	go test ./...

.PHONY: verify-docs
verify-docs:
	go run $(CMD) verify docs

.PHONY: build
build:
	mkdir -p $(DIST)
	go build -o $(DIST)/$(APP) $(CMD)

.PHONY: build-cross
build-cross:
	mkdir -p $(DIST)
	GOOS=darwin GOARCH=arm64 go build -o $(DIST)/$(APP)-darwin-arm64 $(CMD)
	GOOS=linux GOARCH=amd64 go build -o $(DIST)/$(APP)-linux-amd64 $(CMD)

.PHONY: clean
clean:
	rm -rf $(DIST)
