#!/bin/bash

cd PauseCluster && env GOOS=linux GOARCH=amd64 go build -v -o ../build/pausecluster && cd ..