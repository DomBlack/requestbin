run-local:
	ROOT=. REDIS="127.0.0.1:6379" ELASTICSEARCH="127.0.0.1:9200" TCP_PORT=":9999" $(GOPATH)/bin/gin main.go

run-docker:
	docker-compose up

build:
	docker build -t hamstah/requestbin .
