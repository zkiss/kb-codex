.PHONY: setup
setup:
	make -C client setup

.PHONY: compile
compile:
	make -C server compile
	make -C client compile

.PHONY: test
test:
	make -C server test
	make -C client test

.PHONY: server
server:
	make -C server build

.PHONY: client
client:
	make -C client build

.PHONY: build
build: server client
