package main

import (
	"log"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

// Define o Objeto da Conexão
type Conn struct {
	Id     string //identificacao única
	User   string // Usuário do Cliente
	Socket *websocket.Conn
	Rooms  map[string]*Room
}

// Define o objeto sala
type Room struct {
	Name    string
	Members map[string]*Conn
}

// Define our message object
type Message struct {
	Email    string `json:"email"`
	Username string `json:"username"`
	Message  string `json:"message"`
	Event    string `json:"event"`
	Room     string `json:"room"`
}

var (
	// Armazena todas as CONNs pelo ID
	ConnManager = make(map[string]*Conn)
	// Armazena todas as ROOMs pelo Nome
	RoomManager = make(map[string]*Room)
	// channel broadcast
	broadcast = make(chan Message)
	// Atualiza uma conexao HTTP para Web Socket
	upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool { return true },
	}
)

func main() {
	// Serviço para a App WEB
	fs := http.FileServer(http.Dir("./public"))
	http.Handle("/", fs)

	// Configuração da rota do websocket
	http.HandleFunc("/ws", handleConnections)

	// Ouvir mensagens entrantes no channel broadcast
	go handleMessages()

	// Iniciar o servidor na porta 8000 no localhost
	log.Println("ChatGO Iniciado na porta :8000")
	err := http.ListenAndServe(":8000", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

/**
* Função para tratar as novas conexões
 */
func handleConnections(w http.ResponseWriter, r *http.Request) {
	// Upgrade initial GET request to a websocket
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Fatal(err)
	}
	// Make sure we close the connection when the function returns
	//defer ws.Close()

	// UUID unica para o cliente
	id, err := uuid.NewRandom()
	if err != nil {
		log.Fatal(err)
	}

	c := &Conn{
		Socket: ws,
		User:   id.String(), // momentaneamente o username do usuario é o ID dele
		Id:     id.String(),
		Rooms:  make(map[string]*Room),
	}

	//adiciono o as informações de conexão na lista de conexões do servidor
	ConnManager[c.Id] = c
	log.Printf("Nova Conexão - ID: %s", c.Id)

	//se a conexão estiver OK inicio a leitura do socket para obter informações
	if c != nil {
		go c.readSocket()
	}

}

func (c *Conn) readSocket() {

	defer func() {
		c.Socket.Close()
	}()

	//tratar o que o socket recebe
	for {
		var msg Message

		// Ler as mensagens que são enviadas para o socket
		err := c.Socket.ReadJSON(&msg)

		if err != nil {
			log.Printf("error: %v", err)
			//delete(ConnManager, c.Id)
			break
		}
		log.Printf("MSG: %v", msg)
		HandleData(c, msg)
	}
}

/**
* Função para tratar todos os dados recebidos pelo socket
 */
var HandleData = func(c *Conn, msg Message) {

	switch msg.Event {
	case "add":
		//log.Printf("ADD Room")
		sala := NewRoom(msg.Room)
		log.Printf("Sala %s Criada", sala.Name)
		c.ChangeUser(msg.Username)
		c.Join(msg.Room)
	case "join":
		log.Printf("JOIN Room")
		c.ChangeUser(msg.Username)
		c.Join(msg.Room)
	case "change":
		c.ChangeUser(msg.Username)
		log.Printf("Troca Nick")
		//c.Leave(msg.Room)
	case "leave":
		log.Printf("Leave Room")
		c.Leave(msg.Room)
	default:
		broadcast <- msg
	}
}

/**
*	Função para tratar o broadcast das mensagens recebidas, enviando para suas respectivas salas
 */
func handleMessages() {
	for {
		// Pego a informação que está no channel
		msg := <-broadcast
		log.Printf("Recebido uma nova Informacao!")
		//log.Printf("MSG!: %v", msg)

		// obtenho o room que tem como destino a msg
		room := RoomManager[msg.Room]

		//loop criado com o intuito de enviar a mensagem para todas as conexões de uma determinada sala
		for _, client := range room.Members {
			log.Printf("MSG: %v", msg)
			err := client.Socket.WriteJSON(msg)
			if err != nil {
				log.Printf("error: %v", err)
				client.Socket.Close()
				//delete(ConnManager, client.Id)
			}
		}
	}
}

/**
*	Função Join Room
 */
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

// Cria uma nova ROOM.
func NewRoom(name string) *Room {
	if name == "" {
		return nil
	}
	if _, ok := RoomManager[name]; ok {
		return nil
	}
	r := &Room{
		Name:    name,
		Members: make(map[string]*Conn),
	}
	RoomManager[name] = r
	//log.Printf("Salas: %v", RoomManager)
	return r
}

// Change user.
func (c *Conn) ChangeUser(user string) {
	for _, room := range c.Rooms {

		m := Message{
			Email:    room.Name,
			Username: "Servidor",
			Message:  c.User + " mudou para " + user,
			Event:    "msg",
		}

		broadcast <- m
	}

	c.User = user
}

// joined/leave
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
	}

	broadcast <- m
}

func (c *Conn) Leave(name string) {
	room := RoomManager[name]
	delete(room.Members, c.Id)
	delete(ConnManager, c.Id)
	c.Status(name, true)
	//log.Printf("CONEXOES: %v", ConnManager)
}
