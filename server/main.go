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

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

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

	go func() {
		err := http.ListenAndServe(portListen, nil)
		if err != nil {
			log.Fatal("ListenAndServe: ", err)
		}
	}()

	terminate := make(chan os.Signal, 1)
	signal.Notify(terminate, os.Interrupt)
	<-terminate
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
		msg := Message{}
		err = json.Unmarshal([]byte(str), &msg)
		if err != nil {
			log.Printf("ROLO: %s", str)
			fmt.Printf("Erro ao tratar a msg %s\n", err)
			continue
		}
		if msg.Server != idServer {
			log.Printf("MSG Recebida UDP!: %v", msg)
			handleUDP(msg)
		}
	}
}

func handleUDP(msg Message) {
	switch msg.Event {
	case "add":
		sala := NewRoom(msg.Room)
		log.Printf("Sala %s Criada", sala.Name)
		refreshRooms(msg.Room)
	default:
		canalSocket <- msg
	}
}
