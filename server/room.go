package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
)

// Define o objeto sala.
type Room struct {
	Name      string           // Identificador da sala.
	Members   map[string]*Conn // Endereço das conexões conectadas à está sala.
	AddrRoom  *net.UDPAddr     // Endereço do canal multicast da sala.
	emptyRoom chan bool        // Canal de Verificação de Sala vazia
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

// monitRoom disparada ao criar um grupo multicast verifico se há membros na sala
// se tiver eu escuto o grupo multicast, se nao tiver eu deixo de escutar
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
			err = json.Unmarshal([]byte(str), &msg)
			if err != nil {
				log.Printf("ROLO: %s", str)
				fmt.Printf("Erro ao tratar a msg %s\n", err)
				continue
			}
			if msg.Server != idServer {
				log.Printf("MSG Recebida UDP SALA!: %v", msg)
				handleUDP(msg)
			}
		}
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
