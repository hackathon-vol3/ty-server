package server

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"ty-server/internal/database"

	"github.com/gorilla/securecookie"
	"golang.org/x/crypto/bcrypt"
)

type Credentials struct {
	Name     string `json:"name"`
	Password string `json:"password"`
}

var cookieHandler = securecookie.New(
	securecookie.GenerateRandomKey(64),
	securecookie.GenerateRandomKey(32))

func SignupPage(res http.ResponseWriter, req *http.Request) {
	var creds Credentials
	err := json.NewDecoder(req.Body).Decode(&creds)
	if err != nil {
		http.Error(res, "Invalid request body", http.StatusBadRequest)
		log.Printf("json.NewDecoder error: %v", err)
		return
	}

	db := database.ConnectDB()

	var existingUser string
	err = db.QueryRow("SELECT name FROM users WHERE name=?", creds.Name).Scan(&existingUser)
	if err != nil && err != sql.ErrNoRows {
		http.Error(res, "Server error, unable to create your account.", http.StatusInternalServerError)
		log.Printf("db.QueryRow error: %v", err)
		return
	}

	if existingUser != "" {
		http.Error(res, "User already exists", http.StatusConflict)
		log.Printf("User already exists")
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(creds.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(res, "Server error, unable to create your account.", http.StatusInternalServerError)
		log.Printf("bcrypt.GenerateFromPassword error: %v", err)
		return
	}

	_, err = db.Exec("INSERT INTO users(name, password) VALUES(?, ?)", creds.Name, hashedPassword)
	if err != nil {
		http.Error(res, "Server error, unable to create your account.", http.StatusInternalServerError)
		log.Printf("db.Exec error: %v", err)
		return
	}

	res.WriteHeader(http.StatusOK)
	res.Write([]byte("User created!"))
}

func LoginPage(res http.ResponseWriter, req *http.Request) {
	var creds Credentials
	err := json.NewDecoder(req.Body).Decode(&creds)
	if err != nil {
		http.Error(res, "Invalid request body", http.StatusBadRequest)
		log.Printf("json.NewDecoder error: %v", err)
		return
	}

	db := database.ConnectDB()

	var databasePassword string

	err = db.QueryRow("SELECT password FROM users WHERE name=?", creds.Name).Scan(&databasePassword)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(res, "Invalid login credentials", http.StatusUnauthorized)
			log.Printf("User not found")
		} else {
			http.Error(res, "Server error, unable to log you in.", http.StatusInternalServerError)
			log.Printf("db.QueryRow error: %v", err)
		}
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(databasePassword), []byte(creds.Password))
	if err != nil {
		http.Error(res, "Invalid login credentials", http.StatusUnauthorized)
		log.Printf("bcrypt.CompareHashAndPassword error: %v", err)
		return
	}

	SetSession(creds.Name, res)
	res.WriteHeader(http.StatusOK)
	res.Write([]byte("Logged in!"))
}

func SetSession(userName string, res http.ResponseWriter) {
	value := map[string]string{
		"name": userName,
	}
	if encoded, err := cookieHandler.Encode("session", value); err == nil {
		cookie := &http.Cookie{
			Name:   "session",
			Value:  encoded,
			Path:   "/",
			MaxAge: 3600,
		}
		http.SetCookie(res, cookie)
	}
}
