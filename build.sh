#!/bin/bash

for d in `find . -type d -depth 1 ! -name 'example'`
do
    echo "Building $d"
    go build $d
    go list -f '{{join .Imports "\n"}}' $d
done

for f in ./example/*.go
do
    echo "Building $f"
    go build -o ./example/tmp/build $f 
done

rm ./example/tmp/build



