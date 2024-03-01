package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"net"
	"strings"
	"sync"
	"time"
)

type Player struct {
	ID      int
	conn    net.Conn
	message chan string
}

type GameServer struct {
	players  map[int]*Player
	mutex    sync.Mutex
	ctx      context.Context
	cancel   context.CancelFunc
	incoming chan net.Conn
}

func NewGameService() *GameServer {
	ctx, cancel := context.WithCancel(context.Background())
	return &GameServer{
		players:  make(map[int]*Player),
		ctx:      ctx,
		cancel:   cancel,
		incoming: make(chan net.Conn),
	}
}

func (s *GameServer) Run() {
	go func() {
		for {
			select {
			case <-s.ctx.Done():
				time := time.Now()
				log.Printf("game server has done at %v:", time)
				return
			case conn := <-s.incoming:
				go s.HandleConnection(conn)
			}
		}
	}()
}

func (s *GameServer) HandleConnection(conn net.Conn) {
	defer conn.Close()

	player := &Player{
		ID:      len(s.players) + 1,
		conn:    conn,
		message: make(chan string),
	}
	s.mutex.Lock()
	s.players[player.ID] = player
	s.mutex.Unlock()
	fmt.Fprintf(conn, "Welcomme to the game!Your ID is %d\n", player.ID)

	//broadcast new player infomation
	s.Broadcast(fmt.Sprintf("Player %d joined the game\n", player.ID))

	//reade user input
	Scanner := bufio.NewScanner(conn)
	for Scanner.Scan() {
		input := Scanner.Text()
		switch strings.TrimSpace(input) {
		case "quit":
			s.mutex.Lock()
			delete(s.players, player.ID)
			s.mutex.Unlock()
			s.Broadcast(fmt.Sprintf("Player %d left the game\n", player.ID))
			return
		default:
			s.Broadcast(fmt.Sprintf("Player %d: %s\n", player.ID, input))
		}

	}

}

func (s *GameServer) Broadcast(msg string) {
	for _, player := range s.players {
		player.message <- msg
	}
}

func (s *GameServer) Listen(address string) error {
	listener, err := net.Listen("tcp", address)
	if err != nil {
		return err
	}
	defer listener.Close()

	log.Println("Game Server start listening")

	for {
		conn, err := listener.Accept()
		log.Printf("1 connect from %s", conn.RemoteAddr())
		if err != nil {
			fmt.Println("Error accepting", err)
			continue
		}
		s.incoming <- conn
	}
}

func Core_01() {
	server := NewGameService()
	server.Run()

	err := server.Listen(":1234")
	if err != nil {
		log.Println("Error start server:", err)
		return
	}
}
