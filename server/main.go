package main

import (
	//"bufio"
	"encoding/json"
	"flag"

	//"fmt"
	"log"
	//"net"
	"net/http"
	//"os"
	//"os/signal"
	//"runtime/pprof"
	"strconv"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/streadway/amqp"
)

// Define o objeto mensagem.
type message struct {
	Email    string `json:"email"`
	Username string `json:"username"`
	Message  string `json:"message"`
	Event    string `json:"event"`
	Room     string `json:"room"`
	Server   string `json:"server"`
}

// Declaração de variáveis.
var (
	connManager = make(map[string]*conn)       // Armazena todas as CONNs pelo ID.
	roomManager = make(map[string]*room)       // Armazena todas as ROOMs pelo Nome.
	addrManager = make(map[string]*amqp.Queue) // Armazena todos os endereços das Queues pelo nome.
	portManager = make(map[string]string)      // Arnazeba o nome dos grupos multicas pela porta.

	canalSocket = make(chan message) // Canal responsavel pelas mensagens para o websocket.
	canalServer = make(chan message) // Canal responsavel pelas mensagens administrativas entre servidores.

	canalRabbit *amqp.Channel // Canal AMQP do rabit.

	// Menssagens multicast.
	readStr = make([]byte, 1024)

	// Geracao do id de indetificacao de cada servidor.
	id, _    = uuid.NewRandom()
	idServer = id.String()

	// Atualiza uma conexao HTTP para Web Socket.
	upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool { return true },
	}

	portListen string
)

func init() {
	flag.StringVar(&portListen, "port", ":8000", "set the server bind address, e.g.: './server -port :8000'")
}

// Tratamento de erros.
func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}

// Função principal.
func main() {
	flag.Parse()
	forever := make(chan bool)

	// Connecta ao Rabit.
	conn, err := amqp.Dial("amqp://guest:guest@rabitmq:5672/")
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	// Cria o Canal de comunicação com o Rabit.
	canalRabbit, err = conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer canalRabbit.Close()

	// Cria as exchanges para cordenar as queues.
	err = canalRabbit.ExchangeDeclare("sys", "fanout", true, false, false, false, nil)
	failOnError(err, "Failed to open a Exchange")

	// Declara as Queues anexadas ao canal.
	sysQueue, err := canalRabbit.QueueDeclare("", false, false, true, false, nil)
	failOnError(err, "Failed to open a Queue")

	// Binda as Queues com os Exchanges.
	err = canalRabbit.QueueBind(sysQueue.Name, "", "sys", false, nil)
	failOnError(err, "Failed to make a Bind")

	// Prepara o canal para consumir das Queues.
	msgSystem, err := canalRabbit.Consume(sysQueue.Name, "", true, false, false, false, nil)
	failOnError(err, "Failed to open a Consume")

	// Define a sala root no sistema.
	_ = newRoom("root")

	// Goroutines
	go sysWrite(&sysQueue)

	// Serviço para a app web.
	fs := http.FileServer(http.Dir("./public"))
	http.Handle("/", fs)

	// Configuração da rota do websocket.
	http.HandleFunc("/ws", handleConnections)

	// Ouvir as mensagens do canal Socket.
	go handleMessages()

	// UUID do Servidor.
	log.Println("Server UUID: ", idServer, "Porta:", portListen)

	// Escutando requicoes http
	go func() {
		err := http.ListenAndServe(portListen, nil)
		if err != nil {
			log.Fatal("ListenAndServe: ", err)
		}
	}()

	// Ouvindo da fila administrativa.
	go func() {
		for mensagem := range msgSystem {
			m := message{}
			_ = json.Unmarshal([]byte(mensagem.Body), &m)
			switch m.Event {
			case "add":
				_ = newRoom(m.Room)
				refreshRooms(m.Room)
			default:
				canalSocket <- m
			}
		}
	}()

	<-forever
}

// Auxilia na obtencao da porta multicast.
func portGenerator() string {
	number := 9999
	strNumber := "1111"
	port := "0000"
	for ; number > 3000; number-- {
		strNumber = strconv.Itoa(number)

		if _, ok := portManager[strNumber]; ok {
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

// Escrever na fila admnistrativa.
func sysWrite(q *amqp.Queue) {
	for {
		m := <-canalServer
		bolB, _ := json.Marshal(m)

		err := canalRabbit.Publish(
			"sys",  //exchange
			q.Name, //routing key
			false,  //mandatory
			false,  //imediate
			amqp.Publishing{
				ContentType: "text/plain",
				Body:        []byte(bolB),
			})
		failOnError(err, "Failed to publish a message")
	}
}
