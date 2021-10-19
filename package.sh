#!/usr/bin/env bash

mkdir -p dist

GOOS=linux CGO_ENABLED=0 go build -o dist/main github.com/relvacode/lambda-ddns

zip -j dist/function.zip dist/main

