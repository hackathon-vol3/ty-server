package server

import (
	"fmt"
	"log"
	"net/http"
	"sync"
	"ty-server/internal/database"

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
	name string
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
		if game.scores[0] > game.scores[1] {
			game.updateRate(0)
		}
		game.broadcastScore()
		game.viewMyRate(game.clients[0])
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

func (game *GameSession) viewMyRate(client *Client) {
	db := database.ConnectDB()
	defer db.Close()

	var rate int
	err := db.QueryRow("SELECT rate FROM users WHERE name = ?", client.name).Scan(&rate)
	if err != nil {
		log.Printf("Error getting rate: %v", err)
	}

	rateJSON := map[string]int{"rate": rate}
	err = client.conn.WriteJSON(rateJSON)
	if err != nil {
		log.Printf("Error sending rate: %v", err)
	}
}

func HandleConnections(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Fatal("Error upgrading to WebSocket:", err)
	}

	name, err := getNameFromRequest(r)
	if err != nil {
		log.Printf("failed to get name from request: %v", err)
		return
	}

	client := &Client{conn: conn, name: name}

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
			// 対戦相手にどこまで入力できているかを送信
			// 例えば、exampleのaまで入力できている場合、{"progress": exa}を送信
			progress := client.game.sentences[client.game.current][:client.game.position[clientIndex]]
			err := client.game.clients[1-clientIndex].conn.WriteJSON(map[string]string{"opponent_progress": progress})
			if err != nil {
				log.Printf("Error sending progress: %v", err)
			}
			// タイプする文字がなくなったら次のセンテンスへ
			if client.game.position[clientIndex] == len(client.game.sentences[client.game.current]) {
				client.game.scores[clientIndex]++
				client.game.moveToNextSentence()
				// すべてのプレイヤーの位置をリセット
				client.game.position = [2]int{0, 0}
				// 現在のスコアを送信
				client.game.broadcastScore()
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

func (game *GameSession) updateRate(winnerIndex int) {
	for i, client := range game.clients {
		if i == winnerIndex {
			// 勝った場合
			client.updateDbRate(10)
		} else {
			// 負けた場合
			client.updateDbRate(-10)
		}
	}
}

func (c *Client) updateDbRate(rate int) {
	db := database.ConnectDB()
	defer db.Close()

	_, err := db.Exec("UPDATE users SET rate = rate + ? WHERE name = ?", rate, c.name)
	if err != nil {
		log.Printf("Error updating rate: %v", err)
	}

	log.Printf("Updated rate for user %v", c.name)
}

func getNameFromRequest(r *http.Request) (string, error) {
	// セッションクッキーを取得
	cookie, err := r.Cookie("session")
	if err != nil {
		return "", fmt.Errorf("failed to get session cookie: %w", err)
	}

	// クッキーをデコード
	value := make(map[string]string)
	err = cookieHandler.Decode("session", cookie.Value, &value)
	if err != nil {
		return "", fmt.Errorf("failed to decode session cookie: %w", err)
	}

	// ユーザー名を取得
	name, ok := value["name"]
	if !ok {
		return "", fmt.Errorf("no name in session cookie")
	}

	return name, nil
}
