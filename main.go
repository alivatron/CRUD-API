package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
)

var (
	tasks   []task
	counter int
	db      *sql.DB
	err     error
)

func init() {
	godotenv.Load()
}

func main() {
	cfg := mysql.NewConfig()
	cfg.User = os.Getenv("MYSQL_USERNAME")
	cfg.Passwd = os.Getenv("MYSQL_PASSWORD")
	cfg.Net = "tcp"
	cfg.Addr = "127.0.0.1:3306"
	cfg.DBName = "recordings"

	db, err = sql.Open("mysql", cfg.FormatDSN())
	if err != nil {
		log.Fatal("error acured while trying to connect to mysql: ", err)
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
		log.Fatal("error from ListenAndServe by 8080 port: ", err)
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
		err := json.NewDecoder(r.Body).Decode(&task)
		if err != nil {
			w.Write([]byte("something is wrong with request body, please make sure that json format is ok"))
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		counter++
		task.Id = counter
		res, err := db.Exec("INSERT INTO tasks (title, done) VALUES(?, ?)", task.Title, task.Done)
		if aff, _ := res.RowsAffected(); aff == 0 || err != nil {
			w.Write([]byte("bad SQL request"))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(task)
	case http.MethodGet:
		var task task
		rows, err := db.Query("SELECT * FROM tasks")
		if err != nil {
			w.Write([]byte("GET error after sql zapros"))
			w.WriteHeader(http.StatusBadRequest)
			return
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
		id, err := strconv.Atoi(strings.Split(strings.Split(r.URL.Path, "/")[2], "_")[1])
		if err != nil {
			w.Write([]byte("wrong id format in URL path"))
			w.WriteHeader(http.StatusBadRequest)
			return
		}
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
		res, err := db.Exec("DELETE FROM tasks WHERE id = ?", id)
		if err != nil {
			w.Write([]byte("problem accured while trying to execute sql request for DELETE method"))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if aff, _ := res.RowsAffected(); aff == 0 {
			w.Write([]byte("no rows was affected but thats ok maybe?"))
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.WriteHeader(204)
	case http.MethodPut:
		id, err := strconv.Atoi(strings.Split(strings.Split(r.URL.Path, "/")[2], "_")[1])
		if err != nil {
			w.Write([]byte("wrong id in URL path"))
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		var task task
		err = json.NewDecoder(r.Body).Decode(&task)
		if err != nil {
			w.Write([]byte("PUT method error after decoding json data"))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		task.Id = id
		req, err := db.Exec("INSERT INTO tasks (id, title, done) VALUES (?, ?, ?) ON DUPLICATE KEY UPDATE title = VALUES(title), done = VALUES(done);", task.Id, task.Title, task.Done)
		if aff, _ := req.RowsAffected(); aff == 0 || err != nil {
			w.Write([]byte("PUT method error after INSERT sql request"))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		tasks[id] = task
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(task)
	}
}
