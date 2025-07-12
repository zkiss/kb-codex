.PHONY: setup
setup:
	npm install

.PHONY: compile
compile: setup
	npm run build

.PHONY: test
test:
	npm test

.PHONY: build
build: test compile
