APP_NAME := vps-inspector

.PHONY: build test lint run web-build bundle-web release-build

build:
	go build -trimpath -ldflags="-s -w" -o bin/$(APP_NAME) ./cmd/vps-inspector

test:
	go test ./...

lint:
	go vet ./...

run:
	go run ./cmd/vps-inspector

web-build:
	cd web && npm run build

bundle-web: web-build
	rm -rf internal/httpapi/webdist
	mkdir -p internal/httpapi/webdist
	cp -R web/dist/. internal/httpapi/webdist/

release-build: bundle-web
	CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o bin/$(APP_NAME) ./cmd/vps-inspector
