#!/usr/bin/env bash
swag init -g ./b-service/main.go -o ./b-service/docs

swag init -g ./a-service/main.go -o ./a-service/docs
