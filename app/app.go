package main

import (
	"bytes"
	"encoding/json"
	"html/template"
	"io"
	"log"
	"net/http"
)

// Порт
const port string = ":8080"

// Данные для входа
type LoginData struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// Страница авторизации
func serveLoginPage(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("templates/index.html")
	if err != nil {
		http.Error(w, "Template error", http.StatusInternalServerError)
		return
	}
	tmpl.Execute(w, nil)
}

// Страница Профиля
func serveProfilePage(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("templates/profile.html")
	if err != nil {
		http.Error(w, "Template error", http.StatusInternalServerError)
		return
	}
	tmpl.Execute(w, nil)
}

// Запрос на авторизацию
func handleLogin(w http.ResponseWriter, r *http.Request) {
	// Обрабатывать только POST запросы
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var loginData LoginData

	// Чтение JSON
	err := json.NewDecoder(r.Body).Decode(&loginData)
	if err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// Подготовка запроса к другому серверу
	body, err := json.Marshal(loginData)
	if err != nil {
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	// Отправка запроса на другой сервер
	resp, err := http.Post("http://localhost:1337/api/auth", "application/json", bytes.NewBuffer(body))
	if err != nil {
		http.Error(w, "Auth server error", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	// Чтение ответа от другого сервера
	w.Header().Set("Content-Type", "application/json")
	if resp.StatusCode != http.StatusOK {
		// Перенаправление ошибки от другого сервера
		body, _ := io.ReadAll(resp.Body)
		w.WriteHeader(resp.StatusCode)
		w.Write(body)
		return
	}

	// Успешный ответ
	body, err = io.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, "Error reading response", http.StatusInternalServerError)
		return
	}
	w.Write(body)
}

// Запрос на валидацию токена
func verifyToken(w http.ResponseWriter, r *http.Request) {
	// Обрабатывать только POST запросы
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var token string

	err := json.NewDecoder(r.Body).Decode(&token)
	if err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// Подготовка запроса к другому серверу
	body, err := json.Marshal(token)
	if err != nil {
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	// Отправка запроса на другой сервер
	resp, err := http.Post("http://localhost:1337/api/verify", "application/json", bytes.NewBuffer(body))
	if err != nil {
		http.Error(w, "Auth server error", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	// Чтение ответа от другого сервера
	w.Header().Set("Content-Type", "application/json")
	if resp.StatusCode != http.StatusOK {
		// Перенаправление ошибки от другого сервера
		log.Println("Server Error " + (string)(resp.StatusCode))
		body, _ := io.ReadAll(resp.Body)
		w.WriteHeader(resp.StatusCode)
		w.Write(body)
		return
	}
	w.WriteHeader(resp.StatusCode)
	w.Write(body)
}

func Run() {
	log.Println("Сервер запущен на " + port)

	// HTML
	http.HandleFunc("/", serveLoginPage)
	http.HandleFunc("/profile", serveProfilePage)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	// API
	http.HandleFunc("/api/login", handleLogin)
	http.HandleFunc("/api/verify", verifyToken)

	// Запуск сервера (Ctrl + C, чтобы выключить)
	err := http.ListenAndServe(port, nil)
	if err != nil {
		log.Fatal(err)
	}
}
