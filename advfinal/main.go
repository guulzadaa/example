package main

import (
	"log"
	"net/http"
	"os"

	"bookstore/internal/db"

	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load()

	client, mongoDB, err := db.Connect()
	if err != nil {
		log.Fatal(err)
	}
	defer client.Disconnect(db.Bg())

	mux := http.NewServeMux()

	mux.Handle("GET /static/",
		http.StripPrefix("/static/",
			http.FileServer(http.Dir("web/static")),
		),
	)

	RegisterRoutes(mux, mongoDB)

	addr := ":8080"
	if p := os.Getenv("PORT"); p != "" {
		addr = ":" + p
	}

	log.Println("Server started at http://localhost" + addr)
	log.Fatal(http.ListenAndServe(addr, mux))
}
