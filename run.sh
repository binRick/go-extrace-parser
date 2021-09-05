#!/bin/bash
cmd="go build -o extrace-parser && ./extrace-parser --help  >/dev/null && passh ./extrace-parser $@"
eval $cmd
#go run . $@
