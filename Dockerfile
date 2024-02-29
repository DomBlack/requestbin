FROM golang as builder

RUN mkdir -p /workspace
ADD go.mod /workspace
ADD go.sum /workspace
ADD cmd /workspace/cmd
WORKDIR /workspace
RUN go build -o /workspace/requestbin ./cmd/requestbin

FROM gcr.io/distroless/base-debian12

EXPOSE 8080
EXPOSE 8081

ADD ./templates /app/templates
ADD ./static /app/static
ADD ./documents /app/documents
ADD ./GeoLite2-City.mmdb /app/GeoLite2-City.mmdb
ADD ./requestbin.index.config /app/requestbin.index.config
COPY --from=builder /workspace/requestbin /requestbin

CMD ["/requestbin"]
