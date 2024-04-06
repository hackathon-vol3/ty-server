package main

import (
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
    CheckOrigin: func(r *http.Request) bool { return true }, // すべてのオリジンを許可
}

var waitingPlayer *Client // 待機中のプレイヤー
var waitingPlayerLock sync.Mutex // 待機中のプレイヤーを保護するためのロック

// クライアントを表す構造体
type Client struct {
    conn *websocket.Conn
    game *GameSession
}

// ゲームセッションを表す構造体
type GameSession struct {
    clients []*Client
}

// 新しいゲームセッションを作成
func NewGameSession() *GameSession {
    return &GameSession{
        clients: make([]*Client, 0, 2), // 2人プレイヤーのゲームを想定
    }
}

// ゲームセッションにクライアントを追加
func (game *GameSession) join(client *Client) {
    game.clients = append(game.clients, client)
    client.game = game
    if len(game.clients) == 2 {
        // ゲームスタートの処理
        game.start()
    }
}

// ゲームの開始処理
func (game *GameSession) start() {
    for _, client := range game.clients {
        // すべてのクライアントにゲーム開始のメッセージを送信
        go handleGameSession(client)
    }
}

func handleGameSession(client *Client) {
    defer client.conn.Close()
    for {
        _, msg, err := client.conn.ReadMessage()
        if err != nil {
            log.Println("Error reading message:", err)
            return
        }
        fmt.Printf("Received: %s", string(msg))
        // ゲーム進行のロジック (メッセージに基づいた処理)
		client.game.broadcastMessage(client, msg)
    }
}

// ゲームセッション内の他のすべてのクライアントにメッセージを送信
func (game *GameSession) broadcastMessage(sender *Client, msg []byte) {
    for _, client := range game.clients {
        // 送信者以外にメッセージを送信
        if client != sender {
            err := client.conn.WriteMessage(websocket.TextMessage, msg)
            if err != nil {
                log.Printf("Error sending message: %v", err)
                continue
            }
        }
    }
}

func handleConnections(w http.ResponseWriter, r *http.Request) {
    conn, err := upgrader.Upgrade(w, r, nil)
    if err != nil {
        log.Fatal("Error upgrading to WebSocket:", err)
    }

    currentClient := &Client{conn: conn}

    waitingPlayerLock.Lock()
    if waitingPlayer == nil {
        // まだ待機中のプレイヤーがいなければ、現在のプレイヤーを待機させる
        waitingPlayer = currentClient
        waitingPlayerLock.Unlock()
        return // このプレイヤーは次のプレイヤーを待つ
    }

    // 2人目のプレイヤーが見つかったのでゲームセッションを開始
    gameSession := NewGameSession()
    gameSession.join(waitingPlayer)
    gameSession.join(currentClient)
    waitingPlayer = nil // 待機プレイヤーをリセット
    waitingPlayerLock.Unlock()
}

func main() {
    http.HandleFunc("/", handleConnections)
    fmt.Println("Server is running on port 8080...")
    if err := http.ListenAndServe(":8080", nil); err != nil {
        log.Fatal("Error starting server:", err)
    }
}
