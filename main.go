package main

import (
	"log"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type Conn struct {
	Cookie map[string]string
	Socket *websocket.Conn
	Id     string
	//Send   chan []byte
	//Rooms  map[string]*Room
}

// Define our message object
type Message struct {
	Email    string `json:"email"`
	Username string `json:"username"`
	Message  string `json:"message"`
	Event    string `json:"event"`
}

var (
	// Stores all Conn types by their uuid.
	ConnManager = make(map[string]*Conn)
	upgrader    = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool { return true },
	}
)

func main() {
	// Create a simple file server
	fs := http.FileServer(http.Dir("./public"))
	http.Handle("/", fs)

	// Configure websocket route
	http.HandleFunc("/ws", handleConnections)

	// Start listening for incoming chat messages
	//go handleMessages()

	// Start the server on localhost port 8000 and log any errors
	log.Println("http server started on :8000")
	err := http.ListenAndServe(":8000", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

func handleConnections(w http.ResponseWriter, r *http.Request) {
	// Upgrade initial GET request to a websocket
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Fatal(err)
	}
	// Make sure we close the connection when the function returns
	//defer ws.Close()

	// Register our new client
	id, err := uuid.NewRandom()
	if err != nil {
		log.Fatal(err)
	}
	c := &Conn{
		Socket: ws,
		Id:     id.String(),
		//Send:   make(chan []byte, 256),
		//Rooms:  make(map[string]*Room),
	}
	log.Printf("Entrou ID: %v", c)
	ConnManager[c.Id] = c

	if c != nil {

		go c.readSocket()

	}

}

func (c *Conn) readSocket() {

	//tratar o que o socket recebe
	for {
		var msg Message

		// Read in a new message as JSON and map it to a Message object
		err := c.Socket.ReadJSON(&msg)

		if err != nil {
			log.Printf("error: %v", err)
			//delete(clients, ws)
			break
		}

		HandleData(c, msg)
		// Send the newly received message to the broadcast channel
		//broadcast <- msg
		//log.Printf("msg: %s", msg)
	}
}

var HandleData = func(c *Conn, msg Message) {
	log.Printf("msg: %s", msg.Event)
}
