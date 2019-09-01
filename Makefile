BINARY=pipeline-test

default: build

build: 
	@go build -o ./bin/$(BINARY) ./

linux_build: 
	@CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ./bin/$(BINARY) ./

windows_build: 
	@CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o ./bin/$(BINARY).exe ./

docker_build: linux_build
	docker build -t shiguanghuxian/$(BINARY) .

docker_run: docker_build
	docker-compose up --force-recreate

run: build
	@./bin/$(BINARY)

install: build
	@mv ./bin/$(BINARY) $(GOPATH)/bin/$(BINARY)

clean: 
	@rm -f ./bin/$(BINARY)*
	@rm -f ./bin/logs/*

.PHONY: default build linux_build windows_build docker_build docker_run run install clean