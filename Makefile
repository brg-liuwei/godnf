
.PHONY: help test bench clean

VARS = vars.mk


$(shell ./make_config.sh ${VARS})
include ${VARS}

help:
	@echo "Usage: use \`make test\` to run example"
	@echo "Usage: use \`make bench\` to run example and benchmark"

test:
	go test -v -coverprofile=coverage.txt -covermode=atomic

bench:
	go test -bench="." -cpu=${NCPU}

clean:
	rm -rf ${VARS}
