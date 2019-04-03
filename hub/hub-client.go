package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
)

var writeStr, readStr = make([]byte, 1024), make([]byte, 1024)

func main() {
	var (
		host   = "127.0.0.1"
		port   = "32768"
		remote = host + ":" + port
		reader = bufio.NewReader(os.Stdin)
	)

	con, err := net.Dial("tcp", remote)
	defer con.Close()

	if err != nil {
		fmt.Println("Server not found.")
	}

	go read(con)

	for {
		writeStr, _, _ = reader.ReadLine()

		in, err := con.Write([]byte(writeStr))
		if err != nil {
			fmt.Printf("Error when send to server: %d\n", in)
		}

	}
}

func read(conn net.Conn) {
	for {
		length, err := conn.Read(readStr)
		if err != nil {
			fmt.Printf("Error when read from server. Error:%s\n", err)
		}
		fmt.Println(string(readStr[:length]))
	}
}
