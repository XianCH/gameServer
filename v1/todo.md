```go
package main

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"strings"
	"sync"
)

// Player 表示一个游戏玩家
type Player struct {
	ID      int
	conn    net.Conn
	message chan string
}

// GameServer 表示游戏服务器
type GameServer struct {
	players  map[int]*Player
	mutex    sync.Mutex
	ctx      context.Context
	cancel   context.CancelFunc
	incoming chan net.Conn
}

// NewGameServer 创建一个新的游戏服务器
func NewGameServer() *GameServer {
	ctx, cancel := context.WithCancel(context.Background())
	return &GameServer{
		players:  make(map[int]*Player),
		ctx:      ctx,
		cancel:   cancel,
		incoming: make(chan net.Conn),
	}
}

// Run 启动游戏服务器
func (s *GameServer) Run() {
	go func() {
		for {
			select {
			case <-s.ctx.Done():
				return
			case conn := <-s.incoming:
				go s.HandleConnection(conn)
			}
		}
	}()
}

// HandleConnection 处理玩家连接
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

	fmt.Fprintf(conn, "Welcome to the game! Your ID is %d\n", player.ID)

	// 广播新玩家加入的消息
	s.Broadcast(fmt.Sprintf("Player %d joined the game\n", player.ID))

	// 读取玩家输入
	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		input := scanner.Text()

		// 处理玩家命令
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

	// 玩家断开连接
	s.mutex.Lock()
	delete(s.players, player.ID)
	s.mutex.Unlock()
	s.Broadcast(fmt.Sprintf("Player %d left the game\n", player.ID))
}

// Broadcast 广播消息给所有玩家
func (s *GameServer) Broadcast(msg string) {
	for _, player := range s.players {
		player.message <- msg
	}
}

// Listen 启动服务器监听
func (s *GameServer) Listen(address string) error {
	listener, err := net.Listen("tcp", address)
	if err != nil {
		return err
	}
	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error accepting:", err)
			continue
		}
		s.incoming <- conn
	}
}

func main() {
	server := NewGameServer()
	server.Run()

	err := server.Listen(":1234")
	if err != nil {
		fmt.Println("Error starting server:", err)
		return
	}
}

```
