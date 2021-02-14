#!/bin/bash

if [[ $1 == "watch" ]]; then
    ls ./examples/**/*.yml ./templates/*.tmpl ./public/site.css ./*.go | \
        entr docker run -it -v `pwd`:/go/src/app/ ghactionsbyexample
else
    docker build . -t ghactionsbyexample && docker run -it -v `pwd`/public:/go/src/app/public/ ghactionsbyexample
fi
