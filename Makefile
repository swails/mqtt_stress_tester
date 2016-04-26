all: build

deps:
	go get github.com/eclipse/paho.mqtt.golang
	go get github.com/montanaflynn/stats

build:
	go build stresstester

test:
	/bin/rm -f broker/mosquitto.log
	./run_all_tests.sh
