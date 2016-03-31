package main

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/garyburd/redigo/redis"
	"github.com/satori/go.uuid"
	"gopkg.in/olivere/elastic.v3"
)

func ListBins(redisClient redis.Conn) []string {
	bins, err := redis.Strings(redisClient.Do("SMEMBERS", "bins"))
	if err != nil {
		panic(err)
	}
	return bins
}

func ListRequestsFromBin(redisClient redis.Conn, binId string) []HttpRequest {
	raw_requests, err := redis.Strings(redisClient.Do("LRANGE", "bins:"+binId, 0, 10))
	if err != nil {
		panic(err)
	}
	var requests = make([]HttpRequest, len(raw_requests))
	for i, item := range raw_requests {
		if err = json.Unmarshal([]byte(item), &requests[i]); err != nil {
			panic(err)
		}
	}
	return requests
}

func StoreRequest(redisClient redis.Conn, elasticsearchClient *elastic.Client, binId string, request HttpRequest) {
	serialised, err := json.Marshal(request)
	if err != nil {
		panic(err)
	}
	binKey := "bins:" + binId
	if _, err := redisClient.Do("SADD", "bins", binId); err != nil {
		fmt.Println(err)
	}
	if _, err := redisClient.Do("LPUSH", binKey, string(serialised)); err != nil {
		fmt.Println(err)
	}
	if _, err := redisClient.Do("EXPIRE", binKey, 3600*24); err != nil {
		fmt.Println(err)
	}

	record := struct {
		Request HttpRequest `json:"request"`
		Time    time.Time   `json:"time"`
		BinId   string      `json:"bin_id"`
	}{
		Request: request,
		Time:    request.Time,
		BinId:   binId,
	}

	_, err = elasticsearchClient.Index().
		Index("requestbin").
		Type("http").
		BodyJson(record).
		Id(uuid.NewV4().String()).
		Do()
	if err != nil {
		fmt.Println(err)
	}

}
