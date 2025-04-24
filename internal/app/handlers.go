package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
)

type TaskHandler struct {
	DB *sql.DB
}
type Task struct {
	Id    int    `json:"id"`
	Title string `json:"title"`
	Done  bool   `json:"done"`
}

func (h *TaskHandler) Handle(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		var task Task
		err := json.NewDecoder(r.Body).Decode(&task)
		if err != nil {
			w.Write([]byte("bad request: " + err.Error()))
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		res, err := h.DB.Exec("INSERT INTO tasks (title, done) VALUES(?, ?)", task.Title, task.Done)
		if aff, _ := res.RowsAffected(); aff == 0 || err != nil {
			w.Write([]byte("bad SQL request"))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(task)
	case http.MethodGet:
		var task Task
		rows, err := h.DB.Query("SELECT * FROM tasks")
		if err != nil {
			w.Write([]byte("GET error after sql zapros"))
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		var tasks []Task
		for rows.Next() {
			rows.Scan(&task.Id, &task.Title, &task.Done)
			tasks = append(tasks, task)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(tasks)
	}
}

func (h *TaskHandler) HandleId(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		var task Task
		id, err := strconv.Atoi(strings.Split(strings.Split(r.URL.Path, "/")[2], "_")[1])
		if err != nil {
			w.Write([]byte("wrong id format in URL path"))
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		row := h.DB.QueryRow("SELECT id, title, done FROM tasks WHERE id = ?", id)
		if err = row.Scan(&task.Id, &task.Title, &task.Done); err != nil {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte("no matches in table"))
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(task)
	case http.MethodDelete:
		id, _ := strconv.Atoi(strings.Split(strings.Split(r.URL.Path, "/")[2], "_")[1])
		res, err := h.DB.Exec("DELETE FROM tasks WHERE id = ?", id)
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
		var task Task
		err = json.NewDecoder(r.Body).Decode(&task)
		if err != nil {
			w.Write([]byte("PUT method error after decoding json data"))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		task.Id = id
		req, err := h.DB.Exec("INSERT INTO tasks (id, title, done) VALUES (?, ?, ?) ON DUPLICATE KEY UPDATE title = VALUES(title), done = VALUES(done);", task.Id, task.Title, task.Done)
		if aff, _ := req.RowsAffected(); aff == 0 || err != nil {
			w.Write([]byte("PUT method error after INSERT sql request"))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(task)
	}
}
