package main

import (
	"auction/internal/handlers"
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"log"
	"net/http"
)

func main() {
	// Инициализация базы данных
	post := "user=postgres password=Ambb5xh5dr6ss dbname=auction host=localhost port=5432 sslmode=disable"
	db, err := sql.Open("postgres", post)
	if err != nil {
		fmt.Printf("error open db", err)
	}
	defer db.Close()

	// Инициализация обработчиков
	authHandler := handlers.NewAuthHandler(db)
	lotHandler := handlers.NewLotHandler(db)
	bidHandler := handlers.NewBidHandler(db)

	// Настройка маршрутов
	http.HandleFunc("/api/register", authHandler.Register)
	http.HandleFunc("/api/login", authHandler.Login)
	http.HandleFunc("/api/logout", authHandler.Logout)

	http.HandleFunc("/api/lots", lotHandler.GetLots)
	http.HandleFunc("/api/lots/create", lotHandler.CreateLot)
	http.HandleFunc("/api/lots/", lotHandler.GetLot)

	http.HandleFunc("/api/bids/create", bidHandler.CreateBid)
	http.HandleFunc("/api/bids/my", bidHandler.GetMyBids)

	log.Println("The server is running at :8081")
	log.Fatal(http.ListenAndServe(":8081", nil))
}
