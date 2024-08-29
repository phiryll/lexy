#!/bin/bash

# This script is for running benchmark tests during development.
# A typical use would look like:
#
#   $ ./bench.sh Float64 3
#
# Or if "Float64" unintentionally matches multiple benchmarks:
#
#   $ ./bench.sh "BencharkFloat64$" 3

set -e

cd "$(dirname "$0")"
rm -f lexy.test
go test -c

tests=${1:-.}
count=${2:-10}
go test -bench "${tests}" -benchmem -timeout 0 -count=${count}
