package main

import (
	"encoding/json"
	"fmt"

	"github.com/garyburd/redigo/redis"
	"github.com/oschwald/geoip2-golang"
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

type HttpRequestWriter interface {
	WriteHttpRequest(request HttpRequest) error
}

type TcpRequestWriter interface {
	WriteTcpRequest(request TcpRequest) error
}

type RedisHttpRequestWriter struct {
	client redis.Conn
}

func (w RedisHttpRequestWriter) WriteHttpRequest(request HttpRequest) error {
	serialised, err := json.Marshal(request)
	if err != nil {
		panic(err)
	}
	binKey := "bins:" + request.BinId
	if _, err := w.client.Do("SADD", "bins", request.BinId); err != nil {
		fmt.Println(err)
	}
	if _, err := w.client.Do("LPUSH", binKey, string(serialised)); err != nil {
		fmt.Println(err)
	}
	if _, err := w.client.Do("EXPIRE", binKey, 3600*24); err != nil {
		fmt.Println(err)
	}
	return nil
}

type ElasticsearchRequestWriter struct {
	client  *elastic.Client
	GeoIPDB *geoip2.Reader
}

func (w ElasticsearchRequestWriter) WriteJSONRequest(requestType string, request interface{}) error {
	_, err := w.client.Index().
		Index("requestbin").
		Type(requestType).
		BodyJson(request).
		Id(uuid.NewV4().String()).
		Do()
	if err != nil {
		fmt.Println("Failed to save to ElasticSearch")
		fmt.Println(err)
	}
	return err
}

func (w ElasticsearchRequestWriter) WriteTcpRequest(request TcpRequest) error {
	lat, lon, _ := RemoteAddrToGeoIP(w.GeoIPDB, request.RemoteAddr)
	record := struct {
		Request  interface{}      `json:"request"`
		Location elastic.GeoPoint `json:"location"`
	}{Request: request, Location: *elastic.GeoPointFromLatLon(lat, lon)}
	return w.WriteJSONRequest("tcp", record)
}

func (w ElasticsearchRequestWriter) WriteHttpRequest(request HttpRequest) error {
	lat, lon, _ := RemoteAddrToGeoIP(w.GeoIPDB, request.RemoteAddr)
	record := struct {
		Request  interface{}      `json:"request"`
		Location elastic.GeoPoint `json:"location"`
	}{Request: request, Location: *elastic.GeoPointFromLatLon(lat, lon)}
	return w.WriteJSONRequest("http", record)
}
