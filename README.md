# ty-server

オンライン対戦タイピングゲーム

以下でサーバー、データベースを立ち上げる
```bash
make up
```
終了
```bash
make down
```

## 認証
ユーザー認証は、セッションベース認証で行う。

ユーザーがログインすると、1時間のセッション情報がCookieに保存される
### エンドポイント
リクエストボディは以下のようなJSON形式を想定
```json
{
    "name": "masa",
    "password": "passw0rd"
}
```
- ユーザー登録
    - **POST:localhost:8080/signup**
    - レスポンス
        - OKの場合
            - `User created!`
            - StatusCode: 200
        - リクエストが不正な場合
            - `Invalid request body`
            - StatusCode: 400
        - ユーザーがすでに存在する場合
            - `User already exists`
            - StatusCode: 409
        - それ以外のエラー
            - StatusCode: 500
- ログイン
    - **POST:localhost:8080/login**
    - レスポンス
        - OKの場合
            - `Logged in!`
            - StatusCode: 200
        - リクエストが不正な場合
            - `Invalid request body`
            - StatusCode: 400
        - ユーザー名、またはパスワードが間違っている場合
            - `Invalid login credentials`
            - StatusCode: 401
        - それ以外のエラー
            - StatusCode: 500



## 処理の流れ
- WebSocketで接続
- 2人からの接続があればセッションを開始
- クライアントからプレイヤーのタイピング情報を受け取り、正誤判定を行う
- リアルタイムで入力状況を相手プレイヤーに送信する
- 片方が完了したら次の問題に遷移
    - スコアをデータベースに保存
- 全て終了したら最終スコアを表示

## フロントで行うこと
- スタート画面の表示
- ユーザーからの入力を取得する
- 正誤判定の結果表示
    - サーバーから送られてくる情報をUIにする
         - 間違ったら赤くするとか
- 対戦相手の入力状況を表示
- ゲームの進行状況の表示
    - 現在の問題・スコアなど
- 結果画面の表示
