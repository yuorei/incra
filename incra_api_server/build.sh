#!/bin/bash

mkdir -p ../infra/environments/prod/lambda

# API Server Lambda
GOOS=linux GOARCH=amd64 go build -tags lambda.norpc -o bootstrap main.go
zip bootstrap.zip bootstrap
rm bootstrap
mv bootstrap.zip ../infra/environments/prod/lambda/

# Reminder Lambda
GOOS=linux GOARCH=amd64 go build -tags lambda.norpc -o bootstrap ./cmd/reminder/main.go
zip reminder.zip bootstrap
rm bootstrap
mv reminder.zip ../infra/environments/prod/lambda/
