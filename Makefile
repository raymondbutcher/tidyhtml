.Phony: test

tests: $(shell find . -type f)
	go test -v
