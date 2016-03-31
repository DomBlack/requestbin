package main

import (
	"bufio"
	"bytes"
	"fmt"
	"net"
	"time"

	"github.com/satori/go.uuid"
	"gopkg.in/olivere/elastic.v3"
)

func startTCPServer(elasticsearchClient *elastic.Client) {
	fmt.Println("Starting TCP server on port 9999")
	server, err := net.Listen("tcp", ":9999")

	if server == nil {
		panic(fmt.Sprintf("couldn't start listening: %s", err))
	}
	conns := clientConns(server)
	for {
		go handleConn(<-conns, elasticsearchClient)
	}
}

func clientConns(listener net.Listener) chan net.Conn {
	ch := make(chan net.Conn)
	i := 0
	go func() {
		for {
			client, err := listener.Accept()
			if client == nil {
				fmt.Printf(fmt.Sprintf("couldn't accept: %s", err))
				continue
			}
			i++
			fmt.Printf("%d: %v <-> %v\n", i, client.LocalAddr(), client.RemoteAddr())
			client.SetReadDeadline(time.Now().Add(4 * time.Second))
			ch <- client
		}
	}()
	return ch
}

func handleConn(client net.Conn, elasticsearchClient *elastic.Client) {
	b := bufio.NewReader(client)
	var res bytes.Buffer

	buf := make([]byte, 32)
	for {
		size, err := b.Read(buf)
		res.Write(buf[:size])
		if err != nil {
			break
		}
	}
	fmt.Println("read: " + res.String())
	record := struct {
		Content string    `json:"content"`
		Time    time.Time `json:"time"`
	}{
		Content: res.String(),
		Time:    time.Now(),
	}

	_, err := elasticsearchClient.Index().
		Index("requestbin").
		Type("tcp").
		BodyJson(record).
		Id(uuid.NewV4().String()).
		Do()
	if err != nil {
		fmt.Println(err)
	}
	client.Close()
}
