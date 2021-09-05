#!/bin/bash
cmd="go build -o extrace-parser && ./extrace-parser --help  >/dev/null && passh ./extrace-parser parse /var/log/extrace.log $@"
eval $cmd
#go run . $@
