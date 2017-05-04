
.PHONY: help test bench clean

VARS = vars.mk


$(shell ./make_config.sh ${VARS})
include ${VARS}

help:
	@echo "Usage: use \`make test\` to run example"
	@echo "Usage: use \`make bench\` to run example and benchmark"

# use `make cov` to generate `coverage.txt` for auto check coverage supported by codecov
cov:
	echo "" > coverage.txt
	go test -bench="." -cpu=${NCPU} -v -coverprofile=profile.out -covermode=atomic
	cat profile.out >> coverage.txt
	rm profile.out
	go test github.com/brg-liuwei/godnf/set -bench="." -cpu=${NCPU} -v -coverprofile=profile.out -covermode=atomic
	cat profile.out >> coverage.txt
	rm profile.out

test:
	go test -race github.com/brg-liuwei/godnf
	go test -race github.com/brg-liuwei/godnf/set

bench:
	go test -bench="." -cpu=${NCPU}
	go test github.com/brg-liuwei/godnf/set -bench="." -cpu=${NCPU}

clean:
	rm -rf ${VARS} coverage.txt profile.out
