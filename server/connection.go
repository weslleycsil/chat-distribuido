package main

import (
	"log"
	"strings"

	"github.com/gorilla/websocket"
)

// Define o Objeto da Conexão.
type conn struct {
	Id     string           // Identificação única.
	User   string           // Nickname / Usuário do Cliente.
	Socket *websocket.Conn  // Endereço da conexão.
	Rooms  map[string]*room // Endereço das salas à que pertence.
}

// Leitura de informações pelo socket.
func (c *conn) readSocket() {

	defer func() {
		//desconetar o usuario de todas as salas que ele tinha entrado
		for _, room := range c.Rooms {
			c.Leave(room.Name)
		}
		c.Socket.Close()
	}()

	// Tratar o que o socket recebe.
	for {
		var msg message

		// Ler as mensagens que são enviadas para o socket.
		err := c.Socket.ReadJSON(&msg)

		if err != nil {
			log.Printf("error: %v", err)
			break
		}
		handleData(c, msg)
	}
}

// Entrando nas Rooms
func (c *conn) Join(name string) {
	var room *room

	if _, ok := roomManager[name]; ok {
		room = roomManager[name]
	} else {
		log.Printf(" Sala não existe")
	}
	room.Members[c.Id] = c
	connManager[c.Id].Rooms[name] = room
	c.Status(name, false)
}

// Remove usuario da sala
func (c *conn) Leave(name string) {
	room := roomManager[name]
	c.Status(name, true)
	delete(room.Members, c.Id)
	delete(connManager, c.Id)
	if len(room.Members) <= 0 {
		room.EmptyRoom <- true
	}
}

// Troca username ou nickname.
func (c *conn) ChangeUser(user string) {
	for _, room := range c.Rooms {

		m := message{
			Email:    "email",
			Username: "Servidor",
			Message:  c.User + " mudou para " + user,
			Event:    "msg",
			Room:     room.Name,
			Server:   idServer,
		}

		canalServer <- m
	}
	c.User = user
}

// Avisos do sistema [joined/leave]
func (c *conn) Status(name string, s bool) {
	user := c.User
	action := "Entrou."

	if s {
		action = "Saiu."
	}

	m := message{
		Email:    "email",
		Username: "Servidor",
		Message:  user + " " + action,
		Event:    "msg",
		Room:     name,
		Server:   idServer,
	}

	canalServer <- m
}

func (c *conn) sendList(lista []string, event string) {
	justString := strings.Join(lista, ",")
	m := message{
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
