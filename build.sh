#!/bin/bash

lambda_name="CastAI-PauseClusterOnSchedule"

cd pause-resume && env GOOS=linux GOARCH=amd64 go build -v -o ../build/pause-resume-cluster -ldflags="-s -w" && cd ..
zip -r -j ./build/lambda.zip ./build/pause-resume-cluster 
aws lambda update-function-code --function-name $lambda_name --zip-file fileb://./build/lambda.zip