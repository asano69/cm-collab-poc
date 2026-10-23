.PHONY: all init deps-front build-front build server dev-back test clean

include myapp.env
export

BINARY := myapp

# Ports used by the dev servers (frontend, backend, and PocketBase-style API)
PORTS := 3002 3003 3004

all: # (*) Build frontend assets and start the server
	go run ./cmd/$(BINARY) superuser upsert admin@mail.internal password --dir=pb_data
	go run ./cmd/$(BINARY) serve

deps-update:
	cd frontend && bun add github:asano69/cm-websocket
	GOPROXY=direct go get github.com/asano69/cmgo@main

.PHONY: deps-front
deps-front:
	cd frontend && bun install

.PHONY: build-front
build-front: deps-front
	cd frontend && bun run build

.PHONY: build
build: build-front
	go build -ldflags="-X github.com/asano69/myapp/internal/version.Version=$(VERSION)" -o $(BINARY) ./cmd/$(BINARY)

.PHONY: relay
relay: 
	go run ./cmd/$(BINARY) relay --upstream ws://localhost:3002 --addr 127.0.0.1:3004


.PHONY: clean
	rm -fr ./tmp/ # air
	rm -fr frontend/node_modules/.vite

# port: 3003
.PHONY: dev-front
dev-front: clean deps-update
	bunx concurrently -n "frontend,backend" -c "blue,green" "cd frontend && bun dev" "go run ./cmd/$(BINARY) serve"

# port: 3002
.PHONY: dev-back
dev-back: clean deps-update
	bunx concurrently -n "frontend,backend" -c "blue,green" "cd frontend && bun watch" "air"


.PHONY: test
test: test-front test-back

test-back:
	go test ./cmd/... ./internal/... 

test-front:
	cd frontend && bun run test

lint: typecheck
	golangci-lint run
	cd frontend && bun run lint --fix
	cd frontend && bun run lint:css

.PHONY: typecheck
typecheck: 
	cd frontend && bun run typecheck



format:
	cd frontend && bunx prettier --write "src/**/*.{js,jsx,ts,tsx,css}"

# 本番では、後方互換性のために残しておいたほうが良いかも。
migrate-collections:
	ls -1 migrations/*.go | sort | head -n -1 | xargs rm -f
	yes | go run ./cmd/myapp migrate collections
	ls -1 migrations/*.go | sort | head -n -1 | xargs rm -f

install: build
	go install ./cmd/$(BINARY)



