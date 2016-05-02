PREFIX=/usr/local

all: build

deps:
	./build.sh --deps

build: deps
	./build.sh

install: build
	/bin/mv mqtt_stresser $(PREFIX)/bin

test: deps
	/bin/rm -f broker/mosquitto.log
	./run_all_tests.sh
