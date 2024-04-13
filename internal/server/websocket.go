package server

import (
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

var waitingPlayer *Client
var waitingPlayerLock sync.Mutex

type Client struct {
	conn *websocket.Conn
	game *GameSession
}

type GameSession struct {
	clients   []*Client
	sentences []string
	current   int
	scores    [2]int
	position  [2]int
}

func NewGameSession() *GameSession {
	return &GameSession{
		clients:   make([]*Client, 0, 2),
		sentences: []string{"example", "typing", "game"}, // ここにタイピング問題を設定
		current:   0,
	}
}

func (game *GameSession) join(client *Client) {
	game.clients = append(game.clients, client)
	client.game = game
	if len(game.clients) == 2 {
		game.start()
	}
}

func (game *GameSession) start() {
	game.broadcastSentence()
	for _, client := range game.clients {
		go handleGameSession(client)
	}
}

func (game *GameSession) broadcastSentence() {
	sentence := game.sentences[game.current]
	for _, client := range game.clients {
		err := client.conn.WriteJSON(map[string]string{"sentence": sentence})
		if err != nil {
			log.Printf("Error sending sentence: %v", err)
		}
	}
}

func (game *GameSession) moveToNextSentence() {
	game.current++
	if game.current >= len(game.sentences) {
		game.broadcastScore()
		return
	}
	game.broadcastSentence()
}

func (game *GameSession) broadcastScore() {
	for _, client := range game.clients {
		err := client.conn.WriteJSON(map[string]int{"score1": game.scores[0], "score2": game.scores[1]})
		if err != nil {
			log.Printf("Error sending result: %v", err)
		}
	}
}

func HandleConnections(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Fatal("Error upgrading to WebSocket:", err)
	}

	client := &Client{conn: conn}

	waitingPlayerLock.Lock()
	if waitingPlayer == nil {
		waitingPlayer = client
		waitingPlayerLock.Unlock()
		return
	}

	gameSession := NewGameSession()
	gameSession.join(waitingPlayer)
	gameSession.join(client)
	waitingPlayer = nil
	waitingPlayerLock.Unlock()
}

func handleGameSession(client *Client) {
	defer client.conn.Close()
	clientIndex := 0
	if client == client.game.clients[1] {
		clientIndex = 1
	}
	for {
		_, msg, err := client.conn.ReadMessage()
		if err != nil {
			log.Println("Error reading message:", err)
			break
		}
		messageText := string(msg)

		expectedChar := string(client.game.sentences[client.game.current][client.game.position[clientIndex]])
		if messageText == expectedChar {
			// 正しい文字の場合、位置を更新
			client.game.position[clientIndex]++
			client.sendMessage("ok")
			// 自分の位置を更新したら相手に自分がどこまで進んだかを通
			err := client.game.clients[1-clientIndex].conn.WriteJSON(map[string]interface{}{
				"opponent_input": string(messageText),
				"position":       client.game.position[clientIndex],
			})
			if err != nil {
				log.Printf("Error sending position: %v", err)
			}
			// タイプする文字がなくなったら次のセンテンスへ
			if client.game.position[clientIndex] == len(client.game.sentences[client.game.current]) {
				client.game.scores[clientIndex]++
				client.game.moveToNextSentence()
				// すべてのプレイヤーの位置をリセット
				client.game.position = [2]int{0, 0}
			}
		} else {
			// 誤った文字をタイプした場合の処理
			client.sendMessage("fault")
		}
	}
}

func (c *Client) sendMessage(message string) {
	err := c.conn.WriteMessage(websocket.TextMessage, []byte(message))
	if err != nil {
		log.Printf("Error in sending message: %v", err)
	}
}
