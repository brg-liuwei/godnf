
.PHONY: all help test

all: help

help:
	@echo "Usage: use \`make test\` to run example and benchmark"

test:
	go test -bench="."

