#!/bin/sh
set -ex
go build -v ./cmd/miniooni
go build -v ./cmd/jafar
sudo ./QA/$1.py ./miniooni
