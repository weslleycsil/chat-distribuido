package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

// Define o Objeto da Conexão.
type Conn struct {
	Id     string           // Identificação única.
	User   string           // Nickname / Usuário do Cliente.
	Socket *websocket.Conn  // Endereço da conexão.
	Rooms  map[string]*Room // Endereço das salas à que pertence.
}

// Define o objeto sala.
type Room struct {
	Name    string           // Identificador da sala.
	Members map[string]*Conn // Endereço das conexões conectadas à está sala.
}

// Define o objeto mensagem.
type Message struct {
	Email    string `json:"email"`
	Username string `json:"username"`
	Message  string `json:"message"`
	Event    string `json:"event"`
	Room     string `json:"room"`
}

// Declaração de variáveis.
var (
	ConnManager  = make(map[string]*Conn) // Armazena todas as CONNs pelo ID.
	RoomManager  = make(map[string]*Room) // Armazena todas as ROOMs pelo Nome.
	broadcast    = make(chan Message)     // Chanal responsavel pela transmissão das mensagens.
	broadcastTCP = make(chan Message)     // Chanal responsavel pela transmissão das mensagens.

	//tcp entre servidores
	readStr = make([]byte, 1024)

	// Atualiza uma conexao HTTP para Web Socket
	upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool { return true },
	}
)

// Função principal.
func main() {
	//comunicaçao entre servidores
	con, err2 := net.Dial("tcp", "localhost:8081")

	if err2 != nil {
		fmt.Println("Server not found.")
	}
	go tcpWrite(con)
	go tcpRead(con)

	// Serviço para a App WEB.
	fs := http.FileServer(http.Dir("./public"))
	http.Handle("/", fs)

	// Configuração da rota do websocket.
	http.HandleFunc("/ws", handleConnections)

	// Ouvir mensagens que entram no channel broadcast.
	go handleMessages()

	// Iniciar o servidor na porta 8000 no localhost.
	log.Println("ChatGO Iniciado na porta :8000")
	err := http.ListenAndServe(":8000", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

/*
	### Funções para tratar as novas conexões. ###
*/

// Gerencia a parte inicial da conexão.
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
	log.Printf("Nova Conexão - ID: %s", c.Id)

	//se a conexão estiver OK inicio a leitura do socket para obter informações
	if c != nil {
		go c.readSocket()
	}

}

// Leitura de informações pelo socket.
func (c *Conn) readSocket() {

	defer func() {
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
		log.Printf("MSG: %v", msg)
		HandleData(c, msg)
	}
}

/*
	### Funções para tratar as dados no socket. ###
*/

// Gerencia dos eventos.
var HandleData = func(c *Conn, msg Message) {

	switch msg.Event {
	case "add":
		//log.Printf("ADD Room")
		sala := NewRoom(msg.Room)
		log.Printf("Sala %s Criada", sala.Name)
		//enviar para o tcp
		broadcastTCP <- msg
		//
		refreshRooms(msg.Room)
		c.Join(msg.Room)
	case "join":
		log.Printf("Join Room")
		c.Join(msg.Room)
	case "change":
		c.ChangeUser(msg.Username)
		log.Printf("Change User")
	case "leave":
		log.Printf("Leave Room")
		c.Leave(msg.Room)
	default:
		// Esse cliente tem permissão para esse canal?
		if _, ok := c.Rooms[msg.Room]; ok {
			// Envia a msg para o canal.
			broadcast <- msg
			broadcastTCP <- msg
		} else {
			log.Printf("Permissão Negada")
		}
	}
}

/*
	### Funções para tratar o broadcast das mensagens recebidas. ###
*/
func handleMessages() {
	for {
		// Pego a informação que está no channel
		msg := <-broadcast
		log.Printf("Recebido uma nova Informacao!")
		//log.Printf("MSG!: %v", msg)

		// obtenho o room que tem como destino a msg
		room := RoomManager[msg.Room]

		if len(room.Members) > 0 { // se tiver pessoas na sala
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

// Cria uma nova ROOM.
func NewRoom(name string) *Room {
	r := &Room{
		Name:    "root",
		Members: make(map[string]*Conn),
	}
	// A sala ja existe?
	if _, ok := RoomManager[name]; ok {
		//log.Printf("A sala ja existe")
		r = RoomManager[name]
		return r

		// O nome foi setado?
	} else if name == "" {
		RoomManager[name] = r
		return r

	} else {
		r.Name = name
		RoomManager[name] = r
		//log.Printf("Salas: %v", RoomManager)
		return r

	}
}

// Remove usuario da sala
func (c *Conn) Leave(name string) {
	room := RoomManager[name]
	delete(room.Members, c.Id)
	delete(ConnManager, c.Id)
	c.Status(name, true)
	//log.Printf("CONEXOES: %v", ConnManager)
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
		}

		broadcast <- m
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
	}

	broadcast <- m
	broadcastTCP <- m
}

// Avisos do sistema [sala adicionada]
func refreshRooms(name string) {

	m := Message{
		Email:    "email",
		Username: "Servidor",
		Message:  "add sala",
		Event:    "command",
		Room:     name,
	}
	log.Printf("Aviso de nova sala criada")
	broadcast <- m
}

func tcpWrite(conn net.Conn) {
	defer func() {
		conn.Close()
	}()
	for {
		writeStr := <-broadcastTCP
		fmt.Println("WriteStr: %s", writeStr)
		bolB, _ := json.Marshal(writeStr)
		fmt.Println("bolB: %s", string(bolB))
		in, err := conn.Write(bolB)
		if err != nil {
			fmt.Printf("Error when send to server: %d\n", in)
		}

	}
}
func tcpRead(conn net.Conn) {
	for {
		length, err := conn.Read(readStr)
		if err != nil {
			fmt.Printf("Error when read from server. Error:%s\n", err)
		}

		str := string(readStr[:length])
		log.Printf("ROLO: %S", str)
		msg := Message{}
		json.Unmarshal([]byte(str), &msg)
		log.Printf("MSG!: %v", msg)

		handleTcp(msg)
	}
}

func handleTcp(msg Message) {
	log.Printf("MSG!: %v", msg)
	switch msg.Event {
	case "add":
		sala := NewRoom(msg.Room)
		log.Printf("Sala %s Criada", sala.Name)
		refreshRooms(msg.Room)
	default:
		broadcast <- msg
	}
}
