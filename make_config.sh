#!/bin/bash

OUTPUT=$1

if test -z ${OUTPUT}; then
    echo "Usage: $0 <output-file>"
    exit 1
fi

rm -rf ${OUTPUT}
touch ${OUTPUT}

SYSTEM=$(uname -s)
if [[ ${SYSTEM} == "Darwin" ]]; then
    NCPU=$(sysctl -n hw.ncpu)
elif [[ ${SYSTEM} == "Linux" ]]; then
    NCPU=$(cat /proc/cpuinfo 2> /dev/null | grep processor | wc -l)
else
    NCPU=1 # other system: default value
fi

echo "NCPU=${NCPU}" >> ${OUTPUT}
