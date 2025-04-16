package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-sql-driver/mysql"
)

var (
	tasks   []task
	counter int
	db      *sql.DB
	err     error
)

func main() {
	cfg := mysql.NewConfig()
	cfg.User = "root"
	cfg.Passwd = "88005553535"
	cfg.Net = "tcp"
	cfg.Addr = "127.0.0.1:3306"
	cfg.DBName = "recordings"

	db, err = sql.Open("mysql", cfg.FormatDSN())
	if err != nil {
		log.Fatal(err)
	}

	err = db.Ping()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Connected!")

	ServerMux := http.NewServeMux()

	ServerMux.HandleFunc("/tasks", handle)
	ServerMux.HandleFunc("/tasks/{id}", handleId)

	err = http.ListenAndServe(":8080", ServerMux)
	if err != nil {
		fmt.Print("penis")
	}
}

type task struct {
	Id    int    `json:"id"`
	Title string `json:"title"`
	Done  bool   `json:"done"`
}

func handle(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		var task task
		json.NewDecoder(r.Body).Decode(&task)
		counter++
		task.Id = counter
		db.Exec("INSERT INTO tasks (title, done) VALUES(?, ?)", task.Title, task.Done)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(task)
	case http.MethodGet:
		var task task
		rows, err := db.Query("SELECT * FROM tasks")
		if err != nil {
			w.Write([]byte("GET error after sql zapros"))
			w.WriteHeader(http.StatusBadRequest)
		}
		for rows.Next() {
			rows.Scan(&task.Id, &task.Title, &task.Done)
			tasks = append(tasks, task)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(tasks)
	}
}

func handleId(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		var task task
		id, _ := strconv.Atoi(strings.Split(strings.Split(r.URL.Path, "/")[2], "_")[1])
		row := db.QueryRow("SELECT id, title, done FROM tasks WHERE id = ?", id)
		if err = row.Scan(&task.Id, &task.Title, &task.Done); err != nil {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte("no matches in table"))
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(task)
	case http.MethodDelete:
		id, _ := strconv.Atoi(strings.Split(strings.Split(r.URL.Path, "/")[2], "_")[1])
		res, _ := db.Exec("DELETE FROM tasks WHERE id = ?", id)
		if aff, _ := res.RowsAffected(); aff == 0 {
			w.Write([]byte("no rows was affected but thats ok maybe?"))
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.WriteHeader(204)
	case http.MethodPut:
		id, _ := strconv.Atoi(strings.Split(strings.Split(r.URL.Path, "/")[2], "_")[1])
		var task task
		json.NewDecoder(r.Body).Decode(&task)
		task.Id = id
		db.Exec("INSERT INTO tasks (id, title, done) VALUES (?, ?, ?) ON DUPLICATE KEY UPDATE title = VALUES(title), done = VALUES(done);", task.Id, task.Title, task.Done)
		tasks[id] = task
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(task)
	}
}
