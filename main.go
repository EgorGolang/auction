package main

import (
	_ "auction/docs"
	"auction/internal/handlers"
	"auction/internal/middleware"
	"auction/internal/repository"
	"auction/internal/service"
	"database/sql"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	httpSwagger "github.com/swaggo/http-swagger"
	"log"
	"net/http"
	"os"
)

// @title AuctionInfo
// @contact.name AuctionInfo Service
// @contact.url http://test.com
// @contact.email test@test.com
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
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

	lotRepo := repository.NewPostgresLotRepository(db)
	bidRepo := repository.NewPostgresBidRepository(db)
	userRepo := repository.NewPostgresUserRepository(db)

	lotService := service.NewLotService(lotRepo, userRepo)
	bidService := service.NewBidService(bidRepo, lotRepo)

	authHandler := handlers.NewAuthHandler(db)
	lotHandler := handlers.NewLotHandler(db, lotService)
	bidHandler := handlers.NewBidHandler(db, bidService)

	r := mux.NewRouter()

	r.PathPrefix("/swagger/").Handler(httpSwagger.WrapHandler)

	r.HandleFunc("/api/register", authHandler.Register)
	r.HandleFunc("/api/login", authHandler.Login)
	r.HandleFunc("/api/lots", lotHandler.GetLots)
	r.HandleFunc("/api/lot", lotHandler.GetLotByID)

	auth := r.PathPrefix("/auth").Subrouter()
	auth.Use(middleware.AuthMiddleware)

	auth.HandleFunc("/lots/create", lotHandler.CreateLot)
	auth.HandleFunc("/bids/create", bidHandler.CreateBid)
	auth.HandleFunc("/bids/my", bidHandler.GetMyBids)

	auth.HandleFunc("/lot/delete", lotHandler.DeleteLot)

	log.Println("The server is running at :8081")
	log.Fatal(http.ListenAndServe(":8081", r))

}
