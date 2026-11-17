.PHONY: build-ui build-windows build-linux test clean

APP_NAME := optimanager
DIST_DIR := dist
LDFLAGS := -s -w

build-ui:
	npm --prefix ui run build

build-windows: build-ui
	mkdir -p $(DIST_DIR)
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags="$(LDFLAGS)" -o $(DIST_DIR)/$(APP_NAME)-windows-amd64.exe ./cmd/server
	cp scripts/windows/setup-optimanager.ps1 $(DIST_DIR)/setup-optimanager.ps1

build-linux: build-ui
	mkdir -p $(DIST_DIR)
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="$(LDFLAGS)" -o $(DIST_DIR)/$(APP_NAME)-linux-amd64 ./cmd/server

test:
	go vet ./...
	go test ./... -race -count=1

clean:
	rm -rf $(DIST_DIR)
