#!/bin/bash

GOOS=linux GOARCH=amd64 go build -tags lambda.norpc -o bootstrap main.go         
zip bootstrap.zip bootstrap
rm bootstrap
 
mv bootstrap.zip ./terraform/lambda/
