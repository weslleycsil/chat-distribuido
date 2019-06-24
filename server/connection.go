package main

import (
	"log"
	"strings"

	"github.com/gorilla/websocket"
)

// Define o Objeto da Conexão.
type Conn struct {
	Id     string           // Identificação única.
	User   string           // Nickname / Usuário do Cliente.
	Socket *websocket.Conn  // Endereço da conexão.
	Rooms  map[string]*Room // Endereço das salas à que pertence.
}

// Entrando nas Rooms
func (c *Conn) Join(name string) {
	var room *Room

	if _, ok := RoomManager[name]; ok {
		room = RoomManager[name]
	} else {
		log.Printf(" Sala não existe")
	}
	room.Members[c.Id] = c
	ConnManager[c.Id].Rooms[name] = room
	c.Status(name, false)
}

// Remove usuario da sala
func (c *Conn) Leave(name string) {
	room := RoomManager[name]
	c.Status(name, true)
	delete(room.Members, c.Id)
	delete(ConnManager, c.Id)
	if len(room.Members) <= 0 {
		room.emptyRoom <- true
	}
	//c.Socket.Close() //verificar e retirar
}

// Troca username ou nickname.
func (c *Conn) ChangeUser(user string) {
	for _, room := range c.Rooms {

		m := Message{
			Email:    "email",
			Username: "Servidor",
			Message:  c.User + " mudou para " + user,
			Event:    "msg",
			Room:     room.Name,
			Server:   idServer,
		}

		canalSocket <- m
		canalMult <- m
	}
	c.User = user
}

// Avisos do sistema [joined/leave]
func (c *Conn) Status(name string, s bool) {
	//room := RoomManager[name]
	user := c.User
	action := "Entrou."

	if s {
		action = "Saiu."
	}

	m := Message{
		Email:    "email",
		Username: "Servidor",
		Message:  user + " " + action,
		Event:    "msg",
		Room:     name,
		Server:   idServer,
	}

	canalSocket <- m
	canalMult <- m
}

// Leitura de informações pelo socket.
func (c *Conn) readSocket() {

	defer func() {
		//desconetar o usuario de todas as salas que ele tinha entrado
		for _, room := range c.Rooms {
			c.Leave(room.Name)
		}
		c.Socket.Close()
	}()

	// Tratar o que o socket recebe.
	for {
		var msg Message

		// Ler as mensagens que são enviadas para o socket.
		err := c.Socket.ReadJSON(&msg)

		if err != nil {
			log.Printf("error: %v", err)
			//delete(ConnManager, c.Id)
			break
		}
		//log.Printf("MSG: %v", msg)
		handleData(c, msg)
	}
}

func (c *Conn) sendList(lista []string, event string) {
	justString := strings.Join(lista, ",")
	m := Message{
		Email:    "email",
		Username: "Servidor",
		Message:  justString,
		Event:    event,
		Room:     "",
		Server:   idServer,
	}

	err := c.Socket.WriteJSON(m)
	if err != nil {
		log.Printf("error: %v", err)
		c.Socket.Close()
	}
}
