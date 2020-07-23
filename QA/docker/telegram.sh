#!/bin/sh
set -ex
go build -v ./cmd/miniooni
go build -v ./cmd/jafar
sudo ./QA/telegram/telegram.py ./miniooni
