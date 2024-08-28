#!/bin/bash

# This script runs all benchmarks, writing each benchark's results
# to a separate file in the benchmarks/ directory.
# For the BenchmarkFoo function, the file would be named BenchmarkFoo.tmp
# This is intentional to avoid overwriting the baseline BenchmarkFoo.txt file,
# which is checked into version control.
#
# This script is normally only used when establishing a baseline or
# bencharking on a new system/os/....
# The same effect can be achieved for a single benchmark like this:
#
#   $ ./bench.sh "BenchmarkFoo$" 20 > benchmarks/BenchmarkFoo.tmp

# This is what works on my mac, using homebrew versions of grep and sed.
# The BSD versions supplied with mac os don't work the same way.

set -e

cd "$(dirname "$0")"
rm lexy.test
go test -c

files=$(ggrep -r --include='**_test.go' --files-with-matches 'func Bench' .)

for file in ${files}
do
    funcs=$(gsed -nr 's/func (Bench\w+).*/\1/p' $file)
    for func in ${funcs}
    do
        echo "$func in $file"
        go test -bench "${func}$" -benchmem -timeout 0 -count 20 > benchmarks/${func}.tmp
    done
done
