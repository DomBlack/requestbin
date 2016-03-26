run-local:
	ROOT=. REDIS=":6379" $(GOPATH)/bin/gin main.go

run-docker:
	docker-compose up

build:
	docker build -t hamstah/requestbin .
