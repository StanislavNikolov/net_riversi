#!/bin/bash

GOARCH=amd64 GOOS=linux go build -o dist/client_linux
GOARCH=amd64 GOOS=windows go build -o dist/client_windows.exe

