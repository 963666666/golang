package main

import (
	"bufio"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"log"
	"net"
)
func main() {
	cert, err := tls.LoadX509KeyPair("../server.crt", "../server.key")
	if err != nil {
		log.Println(err)
		return
	}
	certBytes, err := ioutil.ReadFile("../client.crt")
	if err != nil {
		panic("Unable to read cert.pem")
	}
	clientCertPool := x509.NewCertPool()
	ok := clientCertPool.AppendCertsFromPEM(certBytes)
	if !ok {
		panic("failed to parse root certificate")
	}
	config := &tls.Config{
		Certificates: []tls.Certificate{cert},
		ClientAuth:   tls.RequireAndVerifyClientCert,
		ClientCAs:    clientCertPool,
	}
	ln, err := tls.Listen("tcp", ":8088", config)
	if err != nil {
		log.Println(err)
		return
	}
	defer ln.Close()
	fmt.Printf("start server success!!!\n")
	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Printf("ln.Accept err: %s\n", err.Error())
			continue
		}
		go handleConn(conn)
	}
}
func handleConn(conn net.Conn) {
	defer conn.Close()
	r := bufio.NewReader(conn)
	for {
		n, err := conn.Write([]byte("world\n"))
		if err != nil {
			log.Println("conn.Write err", n, err)
			return
		}
		msg, err := r.ReadString('\n')
		if err != nil {
			log.Println(err)
			return
		}
		println(msg)
	}
}