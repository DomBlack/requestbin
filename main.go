package main

import (
	"fmt"
	"os"

	"github.com/garyburd/redigo/redis"
	"gopkg.in/olivere/elastic.v3"
)

func main() {
	fmt.Println("starting")

	redisClient, err := redis.Dial("tcp", os.Getenv("REDIS"))
	if err != nil {
		panic(err)
	}
	defer redisClient.Close()

	fmt.Println(os.Getenv("ELASTICSEARCH"))
	elasticsearchClient, err := elastic.NewClient(
		elastic.SetURL("http://" + os.Getenv("ELASTICSEARCH")),
	)
	if err != nil {
		// Handle error
		panic(err)
	}
	exists, err := elasticsearchClient.IndexExists("requestbin").Do()
	if err != nil {
		panic(err)
	}
	if !exists {
		_, err = elasticsearchClient.CreateIndex("requestbin").Do()
		if err != nil {
			// Handle error
			panic(err)
		}
	}

	httpRoot := os.Getenv("ROOT")
	httpPort := os.Getenv("PORT")

	redisWriter := RedisHttpRequestWriter{client: redisClient}
	elasticsearchWriter := ElasticsearchRequestWriter{client: elasticsearchClient}

	startHTTPServer(httpRoot, httpPort, redisClient, redisWriter, elasticsearchWriter)

	tcpPort := os.Getenv("TCP_PORT")
	startTCPServer(tcpPort, elasticsearchClient)
	fmt.Println("started")

}
