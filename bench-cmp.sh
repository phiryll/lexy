#!/bin/bash

# This script compares benchmark results to the baseline.
# The single argument is a prefix of the test name(s) to compare, following "Benchmark".
# It will find matching benchmark results in benchmarks/ other than the baseline.
# A typical use would look like:
#
#   $ ./bench-cmp.sh Float64
#
# Or to match BigInt, BigFloat, and BigRat benchmarks:
#
#   $ ./bench-cmp.sh Big

# This is what works on my mac, using the homebrew version of grep.
# The BSD version supplied with mac os doesn't work the same way.
# benchstat is assumed to be installed in the normal place.

set -e

cd "$(dirname "$0")"
rm -f lexy.test
go test -c

for prefix in "$@"
do
    for file in $(ls benchmarks/Benchmark${prefix}*.txt)
    do
        base_name=${file##*/}
        no_ext=${base_name%.txt}
        test_name=${no_ext#Benchmark}
        cmp_files=$(ggrep --exclude="${base_name}" -l "Benchmark${test_name}" benchmarks/*)
        ~/go/bin/benchstat -filter ".name:${test_name}" $file $cmp_files
    done
done
