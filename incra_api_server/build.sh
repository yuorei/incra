#!/bin/bash

# API Server Lambda
GOOS=linux GOARCH=amd64 go build -tags lambda.norpc -o bootstrap main.go
zip bootstrap.zip bootstrap
rm bootstrap
mv bootstrap.zip ./terraform/lambda/

# Reminder Lambda
GOOS=linux GOARCH=amd64 go build -tags lambda.norpc -o bootstrap ./cmd/reminder/main.go
zip reminder.zip bootstrap
rm bootstrap
mv reminder.zip ./terraform/lambda/
