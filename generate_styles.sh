#!/bin/bash

style=${1:default}
echo "Generating style for ${style}"
pygmentize -S ${style} -f html -a .highlight > public/default.css
