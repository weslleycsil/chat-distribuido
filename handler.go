package main

import (
	"log"
	"net/http"

	"github.com/google/uuid"
)

// Gerencia a parte inicial e criacao de uma nova conexão.
func handleConnections(w http.ResponseWriter, r *http.Request) {
	// Upgrade initial GET request to a websocket
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Fatal(err)
	}
	// Make sure we close the connection when the function returns
	// defer ws.Close()

	// UUID unica para o cliente
	id, err := uuid.NewRandom()
	if err != nil {
		log.Fatal(err)
	}

	// Armazena a conexão
	c := &conn{
		Socket: ws,
		User:   id.String(), // momentaneamente o username do usuario é o ID dele
		Id:     id.String(),
		Rooms:  make(map[string]*room),
	}

	//adiciono o as informações de conexão na lista de conexões do servidor
	connManager[c.Id] = c
	//log.Printf("Nova Conexão - ID: %s", c.Id)

	//se a conexão estiver OK inicio a leitura do socket para obter informações
	if c != nil {
		go c.readSocket()
		c.Join("root")
	}
}

func handleData(c *conn, msg message) {

	msg.Server = idServer

	switch msg.Event {
	case "add":
		_ = newRoom(msg.Room)
		canalServer <- msg
		refreshRooms(msg.Room)
		c.Join(msg.Room)
	case "listUsers":
		c.sendList(listMembers(msg.Room), "listUsers")
	case "listRooms":
		c.sendList(listRooms(), "listRooms")
	case "join":
		c.Join(msg.Room)
	case "change":
		c.ChangeUser(msg.Username)
	case "leave":
		c.Leave(msg.Room)
	default:
		// Esse cliente tem permissão para esse canal?
		if _, ok := c.Rooms[msg.Room]; ok {
			// Envia a fila de mensgens do rabbit.
			roomManager[msg.Room].Channel <- msg
		} else {
			log.Printf("Permissão Negada")
		}
	}
}

func handleMessages() {
	for {
		// Pego a informação que está no channel
		msg := <-canalSocket

		// obtenho o room que tem como destino a msg
		room := roomManager[msg.Room]

		//Loop criado com o intuito de enviar a mensagem para todas as conexões de uma determinada sala
		for _, client := range room.Members {
			// Escrevo a mensagem para aquela conexao de socket
			err := client.Socket.WriteJSON(msg)
			if err != nil {
				client.Socket.Close()
			}
		}
	}
}
