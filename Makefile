run-local:
	ROOT=. REDIS=":6379" ELASTICSEARCH="127.0.0.1:9200" $(GOPATH)/bin/gin main.go

run-docker:
	docker-compose up

build:
	docker build -t hamstah/requestbin .
