#!/bin/sh

go build
./mikser -silent -host=localhost:4001 &
./mikser -silent -host=localhost:4002 &
./mikser -silent -host=localhost:4003 &
./mikser -silent -host=localhost:4004 &

read "wait"
