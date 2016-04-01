package main

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/garyburd/redigo/redis"
	"gopkg.in/olivere/elastic.v3"
)

func setupRedis(config string) redis.Conn {
	redisClient, err := redis.Dial("tcp", config)
	if err != nil {
		panic(err)
	}
	return redisClient
}

func setupElasticsearch(config string) *elastic.Client {
	elasticsearchClient, err := elastic.NewClient(
		elastic.SetURL("http://"+config),
		elastic.SetHealthcheckTimeoutStartup(30*time.Second),
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
	return elasticsearchClient
}

func main() {
	fmt.Println("starting")

	redisClient := setupRedis(os.Getenv("REDIS"))
	defer redisClient.Close()

	elasticsearchClient := setupElasticsearch(os.Getenv("ELASTICSEARCH"))

	elasticsearchWriter := ElasticsearchRequestWriter{client: elasticsearchClient}

	root := os.Getenv("ROOT")
	httpPort, err := strconv.Atoi(os.Getenv("PORT"))
	if err != nil {
		panic(fmt.Sprintf("Invalid port %s", os.Getenv("PORT")))
	}

	redisWriter := RedisHttpRequestWriter{client: redisClient}
	startLoggingHttpServer(httpPort, redisWriter, elasticsearchWriter)
	startAdminHttpServer(httpPort+1, root+"/static/", redisClient)
	startKibanaProxy(httpPort+2, os.Getenv("KIBANA"), root+"/passwd")

	tcpPort := os.Getenv("TCP_PORT")
	startTCPServer(tcpPort, elasticsearchWriter)
	fmt.Println("started")
}
