package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"time"

	"github.com/gomodule/redigo/redis"
	"github.com/olivere/elastic/v7"
	"github.com/oschwald/geoip2-golang"
)

func setupRedis(config string) redis.Conn {
	redisClient, err := redis.Dial("tcp", config)
	if err != nil {
		panic(err)
	}
	return redisClient
}

func setupElasticsearch(ctx context.Context, root string, host string) *elastic.Client {
	elasticsearchClient, err := elastic.NewClient(
		elastic.SetURL("http://"+host),
		elastic.SetHealthcheckTimeoutStartup(30*time.Second),
	)
	if err != nil {
		panic(err)
	}
	exists, err := elasticsearchClient.IndexExists("requestbin").Do(ctx)
	if err != nil {
		panic(err)
	}
	if !exists {

		file, err := ioutil.ReadFile(root + "/requestbin.index.config")
		if err != nil {
			panic(err)
		}
		fmt.Println(string(file))
		_, err = elasticsearchClient.CreateIndex("requestbin").
			BodyString(string(file)).
			Do(ctx)
		if err != nil {
			panic(err)
		}
	}
	return elasticsearchClient
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	fmt.Println("starting")
	root := os.Getenv("ROOT")

	redisClient := setupRedis(os.Getenv("REDIS"))
	defer redisClient.Close()

	db, err := geoip2.Open(root + "/GeoLite2-City.mmdb")
	if err != nil {
		panic(err)
	}
	defer db.Close()

	elasticsearchClient := setupElasticsearch(ctx, root, os.Getenv("ELASTICSEARCH"))

	elasticsearchWriter := ElasticsearchRequestWriter{client: elasticsearchClient, GeoIPDB: db}

	httpPort, err := strconv.Atoi(os.Getenv("PORT"))
	if err != nil {
		panic(fmt.Sprintf("Invalid port %s", os.Getenv("PORT")))
	}

	redisWriter := RedisHttpRequestWriter{client: redisClient}
	writers := []HttpRequestWriter{
		redisWriter,
		elasticsearchWriter,
	}

	startLoggingHttpServer(httpPort, root+"/static/", writers)
	startAdminHttpServer(httpPort+1, root+"/static/", redisClient)
	startKibanaProxy(httpPort+2, os.Getenv("KIBANA"), root+"/passwd")

	tcpPort := os.Getenv("TCP_PORT")
	startTCPServer(tcpPort, elasticsearchWriter)
	fmt.Println("started")
}
