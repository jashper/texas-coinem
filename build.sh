#!/bin/bash

export GOPATH=$HOME/go
go install github.com/jashper/texas-coinem/Server
go install github.com/jashper/texas-coinem/Release/Server
go install github.com/jashper/texas-coinem/Client
go install github.com/jashper/texas-coinem/Release/Client