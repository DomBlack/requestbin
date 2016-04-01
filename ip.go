package main

import (
	"net"
	"strings"
	"time"

	"github.com/oschwald/geoip2-golang"
)

type Request struct {
	RemoteAddr string    `json:"remote_addr"`
	Time       time.Time `json:"time"`
	GeoIP      GeoIP     `json:"geoip"`
}

type GeoIP struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

func RemoteAddrToGeoIP(db *geoip2.Reader, remoteAddr string) (*GeoIP, error) {
	parts := strings.SplitN(remoteAddr, ":", 2)
	ip := net.ParseIP(parts[0])
	record, err := db.City(ip)
	if err != nil {
		return nil, err
	}

	return &GeoIP{Latitude: record.Location.Latitude, Longitude: record.Location.Longitude}, nil
}
