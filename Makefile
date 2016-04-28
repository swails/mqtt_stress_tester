all: build

deps:
	./build.sh --deps

build: deps
	./build.sh

test:
	/bin/rm -f broker/mosquitto.log
	./run_all_tests.sh
