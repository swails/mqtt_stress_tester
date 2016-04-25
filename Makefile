all: build

build:
	go build stresstester

test:
	./run_all_tests.sh
