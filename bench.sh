#!/bin/bash

set -e

count=${1:-20}
tests=${2:-.}
go test -bench ${tests} -benchmem -timeout 0 -count=${count}
