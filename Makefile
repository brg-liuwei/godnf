
.PHONY: help test bench clean

VARS = vars.mk


$(shell ./make_config.sh ${VARS})
include ${VARS}

help:
	@echo "Usage: use \`make test\` to run example"
	@echo "Usage: use \`make bench\` to run example and benchmark"

cov:
	echo "" > coverage.txt
	go test -v -coverprofile=profile.out -covermode=atomic
	cat profile.out >> coverage.txt
	rm profile.out
	go test -bench="." -cpu=${NCPU} -coverprofile=profile.out -covermode=atomic
	cat profile.out >> coverage.txt
	rm profile.out

test:
	go test -v

bench:
	go test -bench="." -cpu=${NCPU}

clean:
	rm -rf ${VARS} coverage.txt
