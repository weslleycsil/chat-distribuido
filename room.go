package main

import (
	"encoding/json"
	"fmt"
	//"log"
	"github.com/streadway/amqp"

)

// Define o objeto sala.
type room struct {
	Name      string           // Identificador da sala.
	Members   map[string]*conn // Endereço das conexões conectadas à está sala.
	AddrRoom  *amqp.Queue      // Endereço do canal multicast da sala.
	EmptyRoom chan bool        // Canal de Verificação de Sala vazia
	Channel   chan message     // Canal dedicado a escutar as mensagens dessa sala.
}

// Cria uma nova ROOM.
func newRoom(name string) *room {
	r := &room{
		Name:      "root",
		Members:   make(map[string]*conn),
		AddrRoom:  nil,
		EmptyRoom: make(chan bool),
		Channel:   make(chan message),
	}
	// A sala ja existe?
	if _, ok := roomManager[name]; ok {
		r = roomManager[name]
		return r

	// O nome não foi setado
	} else if name == "" {
		name = "SemNome"
		r.Name = name
		roomManager[name] = r

		// Cria as exchanges para cordenar as queues.
		err := canalRabbit.ExchangeDeclare(name, "fanout", true, false, false, false, nil)
		failOnError(err, "Failed to open a Exchange")

		// Declara as Queues anexadas ao canal.
		queue, err := canalRabbit.QueueDeclare("", false, false, true, false, nil)
		r.AddrRoom = &queue
		failOnError(err, "Failed to open a Queue")

		// Binda as Queues com os Exchanges.
		err = canalRabbit.QueueBind(queue.Name, "", name, false, nil)
		failOnError(err, "Failed to make a Bind")

		// Prepara o canal para consumir das Queues.
		msgConsumer, err := canalRabbit.Consume(queue.Name, "", true, false, false, false, nil)
		failOnError(err, "Failed to open a Consume")
		
		go r.writeRoom(r.AddrRoom, name, r.Channel)
		go r.monitRoom(msgConsumer)
		return r
	// O nome foi setado	
	} else {
		r.Name = name
		roomManager[name] = r

		// Cria as exchanges para cordenar as queues.
		err := canalRabbit.ExchangeDeclare(name, "fanout", true, false, false, false, nil)
		failOnError(err, "Failed to open a Exchange")

		// Declara as Queues anexadas ao canal.
		queue, err := canalRabbit.QueueDeclare("", false, false, true, false, nil)
		r.AddrRoom = &queue
		failOnError(err, "Failed to open a Queue")

		// Binda as Queues com os Exchanges.
		err = canalRabbit.QueueBind(queue.Name, "", name, false, nil)
		failOnError(err, "Failed to make a Bind")

		// Prepara o canal para consumir das Queues.
		msgConsumer, err := canalRabbit.Consume(queue.Name, "", true, false, false, false, nil)
		failOnError(err, "Failed to open a Consume")
		
		go r.writeRoom(r.AddrRoom, name, r.Channel)
		go r.monitRoom(msgConsumer)
		return r
	}
}

// monitRoom disparada ao criar um grupo multicast verifico se há membros na sala
// se tiver eu escuto o grupo multicast, se nao tiver eu deixo de escutar
func (r *room) monitRoom(c <-chan amqp.Delivery){
	for {
		select {
		case <-r.EmptyRoom:
			return

		default:
			for mensagem := range c {
				m := message{}
				_ = json.Unmarshal([]byte(mensagem.Body), &m)
				canalSocket <- m
			}
		}
	}
}


// Escrever na fila de mensagens.
func (r*room) writeRoom(q *amqp.Queue, ex string, ch chan message) {
	for {
		m := <-ch
		bolB, _ := json.Marshal(m)

		err := canalRabbit.Publish(
			ex,     //exchange
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

// Atualiza a listagem das salas
func listRooms() []string {
	// Numero de salas no sistema.
	tam := len(roomManager)
	ArrayRooms := make([]string, tam)

	for _, room := range roomManager {
		ArrayRooms = append(ArrayRooms, room.Name)
	}
	fmt.Println(ArrayRooms)
	return ArrayRooms
}

// Atualiza a listagem das membros na sala
func listMembers(RoomName string) []string {
	sala := roomManager[RoomName]
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

	m := message{
		Email:    "email",
		Username: "Servidor",
		Message:  "add sala",
		Event:    "command",
		Room:     name,
		Server:   idServer,
	}
	canalServer <- m
}
