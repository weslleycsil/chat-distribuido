package main

import (
	"fmt"
	"net"
)

var ConnManager []net.Conn

func main() {
	var (
		data = make([]byte, 1024)
	)

	//open connection port
	ln, err := net.Listen("tcp", ":8081")
	if err == nil {
		fmt.Println("Initiating server... (Ctrl-C to stop)")
	} else {
		fmt.Printf("Error when listen, Err: %s\n", err)
	}
	defer ln.Close()

	for {
		var res string
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println("Error accepting client: ", err.Error())
		}
		//armazernar a conexão
		ConnManager = append(ConnManager, conn)

		//tratar conexão
		go func(con net.Conn) {
			fmt.Println("New connection: ", con.RemoteAddr())

			//iniciar receber msgs
			for {
				length, err := con.Read(data)
				if err != nil {
					fmt.Printf("Client quit.\n")
					con.Close()
					return
				}
				res = string(data[:length])
				fmt.Println(res)

				//enviar para todos a msg
				notify(con, res)
			}
		}(conn)
	}

}

// Notify other clients
func notify(conn net.Conn, msg string) {
	for _, con := range ConnManager {
		if con.RemoteAddr() != conn.RemoteAddr() {
			con.Write([]byte(msg))
		}
	}
}
