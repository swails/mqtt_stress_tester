#!/bin/sh

# This is a helper script designed to build stress tester and pull in
# dependencies

if [ `python -c "print('$0'[0] == '/')"` = "True" ]; then
  export GOPATH="`dirname $0`"
else
  export GOPATH="`pwd`/`dirname $0`"
fi

echo "Setting GOPATH to $GOPATH"

if [ "$1" = "--deps" ]; then
  echo "Getting dependencies"
	go get github.com/eclipse/paho.mqtt.golang || exit 1
	go get github.com/montanaflynn/stats || exit 1
  go get github.com/jessevdk/go-flags || exit 1
  exit 0
fi

# Build
go build mqtt_stresser
