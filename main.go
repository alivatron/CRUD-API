package main

import (
	"context"
	"database/sql"
	"fmt"
	handlers "hz/internal/app"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
)

var (
	db  *sql.DB
	err error
)

func init() {
	godotenv.Load()
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		exit := make(chan os.Signal, 1)
		signal.Notify(exit, os.Interrupt, syscall.SIGTERM)
		<-exit
		cancel()
	}()

	cfg := mysql.NewConfig()
	cfg.User = os.Getenv("MYSQL_USERNAME")
	cfg.Passwd = os.Getenv("MYSQL_PASSWORD")
	cfg.Net = "tcp"
	cfg.Addr = "127.0.0.1:3306"
	cfg.DBName = "recordings"

	db, err = sql.Open("mysql", cfg.FormatDSN())
	if err != nil {
		log.Fatal("error on connect to mysql: ", err)
	}
	defer db.Close()

	if err = db.Ping(); err != nil {
		log.Fatal(err)
	}
	fmt.Println("Connected!")

	taskHandler := &handlers.TaskHandler{DB: db}

	ServerMux := http.NewServeMux()

	server := &http.Server{
		Addr:    ":8080",
		Handler: ServerMux,
	}

	ServerMux.HandleFunc("/tasks", taskHandler.Handle)
	ServerMux.HandleFunc("/tasks/{id}", taskHandler.HandleId)

	go func() {
		if err = http.ListenAndServe(server.Addr, ServerMux); err != nil {
			log.Fatal("error from ListenAndServe by 8080 port: ", err)
		}
	}()

	<-ctx.Done()
	server.Shutdown(ctx)
}
