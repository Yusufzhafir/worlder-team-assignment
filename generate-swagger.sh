#!/usr/bin/env bash

cd a-service
./generate-swagger.sh

cd ..
cd b-service
./generate-swagger.sh
