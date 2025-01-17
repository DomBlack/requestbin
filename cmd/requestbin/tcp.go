package main

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"net"
	"time"
)

type TcpRequest struct {
	Request
	Content string `json:"content"`
}

func startTCPServer(port string, writers ...TcpRequestWriter) {
	fmt.Println("Starting TCP server on port " + port)
	server, err := net.Listen("tcp", net.JoinHostPort("0.0.0.0", port))

	if server == nil {
		panic(fmt.Sprintf("couldn't start listening: %s", err))
	}
	conns := clientConns(server)
	for {
		go handleConn(<-conns, writers...)
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
			client.SetReadDeadline(time.Now().Add(4 * time.Second))
			ch <- client
		}
	}()
	return ch
}

func handleConn(client net.Conn, writers ...TcpRequestWriter) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

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
	request := TcpRequest{
		Request{
			Time:       time.Now(),
			RemoteAddr: client.RemoteAddr().String(),
		},
		res.String(),
	}
	for _, writer := range writers {
		err := writer.WriteTcpRequest(ctx, request)
		if err != nil {
			fmt.Println(err)
		}
	}

	client.Close()
}
