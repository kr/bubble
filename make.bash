#!/bin/bash

set -eo pipefail

p=`go list`
go install -ldflags "-X $p/build.BUBBLEROOT $PWD"
