.PHONY: server-build
server-build:
	cd server && go build ./...

.PHONY: server-test
server-test:
	cd server && go test ./...

.PHONY: server
server: server-build server-test

.PHONY: client
client:
	cd client && npm install && npm run test && npm run build

.PHONY: build
build: server client
