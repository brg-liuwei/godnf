
.PHONY: all help test

VARS = vars.mk


$(shell ./make_config.sh ${VARS})
include ${VARS}

all: help

help:
	@echo "Usage: use \`make test\` to run example and benchmark"

test:
	go test -bench="." -cpu=${NCPU}

clean:
	rm -rf ${VARS}
