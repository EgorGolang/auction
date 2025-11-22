package main

import (
	"auction/internal/handlers"
	"auction/internal/middleware"
	"database/sql"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"log"
	"net/http"
	"os"
)

func main() {
	godotenv.Load(".env")
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		log.Fatal("JWT_SECRET environment variable not set")
	}
	post := "user=postgres password=Ambb5xh5dr6ss dbname=auction host=localhost port=5432 sslmode=disable"
	db, err := sql.Open("postgres", post)
	if err != nil {
		fmt.Printf("error open db", err)
	}
	defer db.Close()

	authHandler := handlers.NewAuthHandler(db)
	lotHandler := handlers.NewLotHandler(db)
	bidHandler := handlers.NewBidHandler(db)

	r := mux.NewRouter()

	r.HandleFunc("/api/register", authHandler.Register)
	r.HandleFunc("/api/login", authHandler.Login)
	r.HandleFunc("/api/lots", lotHandler.GetLots)
	r.HandleFunc("/api/lot", lotHandler.GetLot)

	auth := r.PathPrefix("/auth").Subrouter()
	auth.Use(middleware.AuthMiddleware)

	auth.HandleFunc("/lots/create", lotHandler.CreateLot)
	auth.HandleFunc("/bids/create", bidHandler.CreateBid)
	auth.HandleFunc("/bids/my", bidHandler.GetMyBids)

	r.HandleFunc("/api/lot/delete", lotHandler.DeleteLot)

	log.Println("The server is running at :8081")
	log.Fatal(http.ListenAndServe(":8081", r))

}
