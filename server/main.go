package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"runtime/pprof"
	"strconv"
	"strings"

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
		HandleData(c, msg)
	}
}

// Define o objeto sala.
type Room struct {
	Name      string           // Identificador da sala.
	Members   map[string]*Conn // Endereço das conexões conectadas à está sala.
	AddrRoom  *net.UDPAddr     // Endereço do canal multicast da sala.
	emptyRoom chan bool
}

// Cria uma nova ROOM.
func NewRoom(name string) *Room {
	r := &Room{
		Name:      "root",
		Members:   make(map[string]*Conn),
		AddrRoom:  nil,
		emptyRoom: make(chan bool),
	}
	// A sala ja existe?
	if _, ok := RoomManager[name]; ok {
		//log.Printf("A sala ja existe")
		r = RoomManager[name]
		return r

		// O nome foi setado?
	} else if name == "" {
		r.AddrRoom = manageMulticastGroup(name)
		RoomManager[name] = r
		go r.monitRoom()
		return r

	} else {
		r.Name = name
		r.AddrRoom = manageMulticastGroup(name)
		RoomManager[name] = r
		go r.monitRoom()
		return r
	}
}

func (r *Room) monitRoom() {

	Addr := r.AddrRoom
	connListen, _ := net.ListenMulticastUDP("udp", nil, Addr)

	for {
		select {
		case <-r.emptyRoom:
			return

		default:
			length, _, err := connListen.ReadFromUDP(readStr)
			if err != nil {
				fmt.Printf("Error when read from server. Error:%s\n", err)
				continue
			}
			str := string(readStr[:length])
			log.Printf("MSG Recebida UDP Sala: %s", str)
			msg := Message{}
			json.Unmarshal([]byte(str), &msg)
			//log.Printf("MSG!: %v", msg)
			if msg.Server != idServer {
				handleUDP(msg)
			}
		}
	}
}

// Define o objeto mensagem.
type Message struct {
	Email    string `json:"email"`
	Username string `json:"username"`
	Message  string `json:"message"`
	Event    string `json:"event"`
	Room     string `json:"room"`
	Server   string `json:"server"`
}

// Declaração de variáveis.
var (
	ConnManager = make(map[string]*Conn)        // Armazena todas as CONNs pelo ID.
	RoomManager = make(map[string]*Room)        // Armazena todas as ROOMs pelo Nome.
	AddrManager = make(map[string]*net.UDPAddr) // Armazena todos os endereços multicast pelo nome do grupo.
	PortManager = make(map[string]string)       // Arnazeba o nome dos grupos multicas pela porta.
	canalSocket = make(chan Message)            // Canal responsavel pelas mensagens locais pro websocket
	canalAdm    = make(chan Message)            // Canal administrativo responsavel por mensagens entre servidores
	canalMult   = make(chan Message)            // Canal responsavel por enviar mensagens para grupo Multicast

	// Geracao do id de indetificacao de cada servidor.
	id, _    = uuid.NewRandom()
	idServer = id.String()

	//	Dados da trasmicao UDP entre servidores.
	readStr = make([]byte, 1024)

	// Atualiza uma conexao HTTP para Web Socket.
	upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool { return true },
	}

	portListen string
	cpuprofile *string
)

func init() {
	flag.StringVar(&portListen, "port", ":8000", "set the server bind address, e.g.: './server -port :9000'")
	cpuprofile = flag.String("cpuprofile", "", "write cpu profile to a file, e.g.: './server -cpuprofile=magano.prof'")
}

// Função principal.
func main() {

	flag.Parse()
	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			log.Fatal("could not create CPU profile: ", err)
		}
		defer f.Close()
		if err := pprof.StartCPUProfile(f); err != nil {
			log.Fatal("could not start CPU profile: ", err)
		}
		defer pprof.StopCPUProfile()
	}

	// Comunicaçao entre servidores atravez de MulticastUDP
	//groupAdm := "224.30.30.30:9999"
	// Addr, _ := net.ResolveUDPAddr("udp", groupAdm)
	Addr := manageMulticastGroup("Servers")
	conn, err2 := net.DialUDP("udp", nil, Addr)
	connListen, err3 := net.ListenMulticastUDP("udp", nil, Addr)

	if err2 != nil {
		fmt.Println("Server not found.")
	}
	if err3 != nil {
		fmt.Println("Server not found Listen.")
	}
	// Prepara a sala Root
	_ = NewRoom("root")

	// Mensagens Entre servidores
	go udpWriteAdm(conn)
	go udpWrite()
	go udpRead(connListen)

	// Serviço para a App WEB.
	fs := http.FileServer(http.Dir("./public"))
	http.Handle("/", fs)

	// Configuração da rota do websocket.
	http.HandleFunc("/ws", handleConnections)

	// Ouvir mensagens que entram no channel canalSocket.
	go handleMessages()
	log.Println("Server UUID: ", idServer, "Porta:", portListen) //mostra o UUID do servidor
	// Iniciar o servidor na porta 8000 no localhost.
	fmt.Println("\ttecnicamente..")
	go func() {
		err := http.ListenAndServe(portListen, nil)
		if err != nil {
			log.Fatal("ListenAndServe: ", err)
		}
	}()

	terminate := make(chan os.Signal, 1)
	signal.Notify(terminate, os.Interrupt)
	<-terminate

	fmt.Println("\t\tSAIU!")
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
	//log.Printf("Nova Conexão - ID: %s", c.Id)

	//se a conexão estiver OK inicio a leitura do socket para obter informações
	if c != nil {
		go c.readSocket()
		c.Join("root")
	}

}

/*
	### Funções para tratar as dados no socket. ###
*/

// Gerencia dos eventos.
func HandleData(c *Conn, msg Message) {

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

/*
	### Funções para tratar o canalSocket das mensagens recebidas. ###
*/
func handleMessages() {
	for {
		// Pego a informação que está no channel
		msg := <-canalSocket
		//log.Printf("Recebido uma nova Informacao!")
		//log.Printf("MSG!: %v", msg)

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

func portGenerator() string {
	number := 9999
	strNumber := "1111"
	port := "0000"
	for ; number > 3000; number-- {
		strNumber = strconv.Itoa(number)

		if _, ok := PortManager[strNumber]; ok {
			// Porta está em uso.
			//log.Printf("porta em uso, busca a proxima")
		} else {
			// Porta ociosa.
			number = 1
			//log.Printf("porta sem uso, usar")
		}
	}
	lenn := len(strNumber)
	for lenn < 4 {
		port = "0" + port
		lenn = len(port)
	}

	return strNumber
}

func manageMulticastGroup(groupName string) *net.UDPAddr {
	if _, ok := AddrManager[groupName]; ok {
		// Já existe esse um canal para esse grupo.
		return AddrManager[groupName]
	} else {
		// Não existe o canal desse grupo.
		base := "224.30.30.30"
		port := portGenerator()
		PortManager[port] = groupName

		groupAddrs := base + ":" + port
		fmt.Println("Sala: ", groupName, "->", groupAddrs)
		Addr, _ := net.ResolveUDPAddr("udp", groupAddrs)
		AddrManager[groupName] = Addr
		return Addr
	}
}

// Atualiza a listagem das salas
func listRooms() []string {
	// Numero de salas no sistema.
	tam := len(RoomManager)
	ArrayRooms := make([]string, tam)

	for _, room := range RoomManager {
		ArrayRooms = append(ArrayRooms, room.Name)
	}
	fmt.Println(ArrayRooms)
	return ArrayRooms
}

// Atualiza a listagem das membros na sala
func listMembers(RoomName string) []string {
	sala := RoomManager[RoomName]
	// Numero de membros naquela sala.
	tam := len(sala.Members)
	ArrayMembers := make([]string, tam)

	for _, member := range sala.Members {
		ArrayMembers = append(ArrayMembers, member.User)
	}
	fmt.Println(ArrayMembers)
	return ArrayMembers
}

// Avisos do sistema [sala adicionada]
func refreshRooms(name string) {

	m := Message{
		Email:    "email",
		Username: "Servidor",
		Message:  "add sala",
		Event:    "command",
		Room:     name,
		Server:   idServer,
	}
	log.Printf("Aviso de nova sala criada")
	canalSocket <- m
}

func udpWriteAdm(conn *net.UDPConn) {
	for {
		writeStr := <-canalAdm
		//fmt.Println("WriteStr: %s", writeStr)
		bolB, _ := json.Marshal(writeStr)
		fmt.Println("bolB:", string(bolB))
		in, err := conn.Write(bolB)
		if err != nil {
			fmt.Printf("Error when send to server: %d\n", in)
		}

	}

}

//escrita nos multicast groups do canal canal Multicast
func udpWrite() {
	for {
		writeStr := <-canalMult
		sala := writeStr.Room
		fmt.Println("Sala UDPWrite", sala)
		conn, _ := net.DialUDP("udp", nil, RoomManager[sala].AddrRoom)
		//fmt.Println("WriteStr: %s", writeStr)
		bolB, _ := json.Marshal(writeStr)
		fmt.Println("bolB:", string(bolB))
		in, err := conn.Write(bolB)
		if err != nil {
			fmt.Printf("Error when send to server: %d\n", in)
		}

	}

}

func udpRead(connListen *net.UDPConn) {
	for {
		length, _, err := connListen.ReadFromUDP(readStr)
		if err != nil {
			fmt.Printf("Error when read from server. Error:%s\n", err)
		}
		str := string(readStr[:length])
		//log.Printf("ROLO: %S", str)
		msg := Message{}
		json.Unmarshal([]byte(str), &msg)
		log.Printf("MSG Recebida UDP!: %v", msg)
		if msg.Server != idServer {
			handleUDP(msg)
		}
	}

}

func handleUDP(msg Message) {
	//log.Printf("MSG!: %v", msg)
	switch msg.Event {
	case "add":
		sala := NewRoom(msg.Room)
		log.Printf("Sala %s Criada", sala.Name)
		refreshRooms(msg.Room)
	default:
		canalSocket <- msg
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

// funcao disparada ao criar um grupo multicast
//verifico se há membros na sala
//se tiver eu escuto o grupo multicast
//se nao tiver eu deixo de escutar
func monitRoom(name string) {
	// r := RoomManager[name]
	// ch := false

	// done := make(chan int)

	// Addr := r.AddrRoom
	// connListen, _ := net.ListenMulticastUDP("udp", nil, Addr)

	// for {
	// 	if len(r.Members) > 0 && ch == false {
	// 		ch = true
	// 		fmt.Println("Monitorar Sala ", r.Name)
	// 		go func() {
	// 			for {
	// 				select {
	// 				case <-done:
	// 					return
	// 				default:
	// 					length, _, err := connListen.ReadFromUDP(readStr)
	// 					if err != nil {
	// 						fmt.Printf("Error when read from server. Error:%s\n", err)
	// 					}
	// 					str := string(readStr[:length])
	// 					log.Printf("MSG Recebida UDP Sala: %s", str)
	// 					msg := Message{}
	// 					json.Unmarshal([]byte(str), &msg)
	// 					//log.Printf("MSG!: %v", msg)
	// 					if msg.Server != idServer {
	// 						handleUDP(msg)
	// 					}
	// 				}
	// 			}
	// 		}()

	// 	}
	// 	if len(r.Members) < 1 && ch == true {
	// 		ch = false
	// 		fmt.Println("Parar Monitoramento da Sala ", r.Name)
	// 		done <- 1
	// 	}
	// 	fmt.Println("AQUIPORR")
	// }
}
