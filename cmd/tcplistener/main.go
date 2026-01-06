package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net"
)

func getLinesChannel(f io.ReadCloser) <-chan string {
	out := make(chan string, 1)
	go func() {
		defer f.Close()
		defer close(out)

		line := ""
		for {
			data := make([]byte, 8)
			n, err := f.Read(data)
			if err != nil {
				break
			}

			data = data[:n]
			if i := bytes.IndexByte(data, '\n'); i != -1 {
				line += string(data[:i])
				data = data[i+1:]
				out <- line
				line = ""
			}

			line += string(data)
		}

		if len(line) != 0 {
			out <- line
		}
	}()

	return out
}

func main() {
	listener, err := net.Listen("tcp", ":4200")
	if err != nil {
		log.Fatal("error: ", err)
	}

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Fatal("error: ", err)
		}
		for line := range getLinesChannel(conn) {
			fmt.Println(line)
		}

	}
}
