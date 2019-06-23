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
	c := &Conn{
		Socket: ws,
		User:   id.String(), // momentaneamente o username do usuario é o ID dele
		Id:     id.String(),
		Rooms:  make(map[string]*Room),
	}

	//adiciono o as informações de conexão na lista de conexões do servidor
	ConnManager[c.Id] = c
	//log.Printf("Nova Conexão - ID: %s", c.Id)

	//se a conexão estiver OK inicio a leitura do socket para obter informações
	if c != nil {
		go c.readSocket()
		c.Join("root")
	}
}

func handleData(c *Conn, msg Message) {

	msg.Server = idServer

	switch msg.Event {
	case "add":
		//log.Printf("ADD Room")
		_ = NewRoom(msg.Room)
		//log.Printf("Sala %s Criada", sala.Name)
		//enviar para o tcp
		canalAdm <- msg
		//
		refreshRooms(msg.Room)
		c.Join(msg.Room)
	case "listUsers":
		log.Printf("Lista de Usuarios")
		c.sendList(listMembers(msg.Room), "listUsers")
	case "listRooms":
		log.Printf("Lista de Salas")
		c.sendList(listRooms(), "listRooms")
	case "join":
		//log.Printf("Join Room")
		c.Join(msg.Room)
	case "change":
		c.ChangeUser(msg.Username)
		//log.Printf("Change User")
	case "leave":
		//log.Printf("Leave Room")
		c.Leave(msg.Room)
	default:
		// Esse cliente tem permissão para esse canal?
		if _, ok := c.Rooms[msg.Room]; ok {
			// Envia a msg para o canal.
			canalSocket <- msg
			canalMult <- msg
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
		room := RoomManager[msg.Room]

		if len(room.Members) > 0 { // se tiver pessoas na sala
			//Loop criado com o intuito de enviar a mensagem para todas as conexões de uma determinada sala
			for _, client := range room.Members {
				log.Printf("MSG: %v", msg)
				// Escrevo a mensagem para aquela conexao de socket
				err := client.Socket.WriteJSON(msg)
				if err != nil {
					log.Printf("error: %v", err)
					client.Socket.Close()
					//delete(ConnManager, client.Id)
				}
			}
		}
	}
}
