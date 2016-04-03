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
}

func RemoteAddrToGeoIP(db *geoip2.Reader, remoteAddr string) (float64, float64, error) {
	parts := strings.SplitN(remoteAddr, ":", 2)
	ip := net.ParseIP(parts[0])
	record, err := db.City(ip)
	if err != nil {
		return 0, 0, err
	}

	return record.Location.Latitude, record.Location.Longitude, nil
}
