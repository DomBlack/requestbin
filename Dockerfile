FROM golang

RUN go get "github.com/garyburd/redigo/redis"
RUN go get "github.com/gorilla/mux"
RUN go get "github.com/jackpal/bencode-go"
RUN go get "github.com/satori/go.uuid"
RUN go get "gopkg.in/olivere/elastic.v3"
RUN go get "github.com/oschwald/geoip2-golang"

ENTRYPOINT /go/bin/requestbin
EXPOSE 8080

ADD ./templates /app/templates
ADD ./static /app/static
ADD ./documents /app/documents
ADD ./GeoLite2-City.mmdb /app/GeoLite2-City.mmdb
ADD ./requestbin.index.config /app/requestbin.index.config

ADD . /go/src/github.com/hamstah/requestbin
RUN go install github.com/hamstah/requestbin
