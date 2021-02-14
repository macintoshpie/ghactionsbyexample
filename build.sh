#!/bin/bash

set -e

if [[ $1 == "watch" ]]; then
    ls examples.txt ./examples/**/*.yml ./templates/*.tmpl ./public/site.css ./*.go | \
        entr docker run -it -v `pwd`:/go/src/app/ ghactionsbyexample
else
    docker build . -t ghactionsbyexample && docker run -it -v `pwd`/public:/go/src/app/public/ ghactionsbyexample

    if [[ $1 == "commit" ]]; then
        echo "Commiting build..."
        ref=$(git rev-parse --abbrev-ref HEAD)
        if [[ $ref == "main" ]]; then
            echo "Can't commit to main branch"
            exit 1
        fi

        git add ./public
        git commit -m "build: site built at $(date)"
    fi
fi
