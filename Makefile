all: build

build:
	go build stresstester

test:
	/bin/rm -f broker/mosquitto.log
	./run_all_tests.sh
