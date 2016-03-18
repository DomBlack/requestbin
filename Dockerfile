FROM golang

RUN go get "github.com/garyburd/redigo/redis"
RUN go get "github.com/gorilla/mux"

ADD . /go/src/github.com/hamstah/requestbin
RUN go install github.com/hamstah/requestbin

ENTRYPOINT /go/bin/requestbin
RUN ls /go/bin
EXPOSE 8080

ADD ./templates /app/templates
ADD ./static /app/static
