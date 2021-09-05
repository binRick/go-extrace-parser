nodemon --delay .3 -w . -e sh,go,sum,yaml,json,ini -I -V -x sh -- -c "./run.sh $@||true"
