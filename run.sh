#!/bin/bash
cmd="go build -o extrace-parser && ./extrace-parser --help && ./extrace-parser log parse /var/log/extrace.log"
eval $cmd
#go run . $@
