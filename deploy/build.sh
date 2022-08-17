#!/bin/bash
target=apigateway
mkdir -p ${target}
cd ..
go build -o deploy/${target}/${target}
cd deploy