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

func (w ElasticsearchRequestWriter) WriteJSONRequest(requestType string, jsonRequest string) error {
	_, err := w.client.Index().
		Index("requestbin").
		Type(requestType).
		BodyJson(jsonRequest).
		Id(uuid.NewV4().String()).
		Do()
	return err
}

func (w ElasticsearchRequestWriter) WriteTcpRequest(request TcpRequest) error {
	geoIP, err := RemoteAddrToGeoIP(w.GeoIPDB, request.RemoteAddr)
	if err == nil {
		request.GeoIP = *geoIP
	}

	jsonRequest, err := json.Marshal(request)
	if err != nil {
		return err
	}
	return w.WriteJSONRequest("tcp", string(jsonRequest))
}

func (w ElasticsearchRequestWriter) WriteHttpRequest(request HttpRequest) error {
	geoIP, err := RemoteAddrToGeoIP(w.GeoIPDB, request.RemoteAddr)
	if err == nil {
		request.GeoIP = *geoIP
	}

	jsonRequest, err := json.Marshal(request)
	if err != nil {
		return err
	}
	return w.WriteJSONRequest("http", string(jsonRequest))
}
