package main

import (
	"log"
	"mongoapi/config"
	"mongoapi/routes"
	"os"
	"net/http"

	"github.com/gorilla/handlers"
)

func main() {
	config.ConnectMongo()

	router := routes.SetupRoutes()

	// Configurar CORS
	headersOk := handlers.AllowedHeaders([]string{"Content-Type"})
	originsOk := handlers.AllowedOrigins([]string{"http://localhost:5173"})
	methodsOk := handlers.AllowedMethods([]string{"GET", "POST", "OPTIONS"})

	log.Println("Servidor iniciado en :8080")
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080" // Puerto por defecto para pruebas locales
	}
	log.Fatal(http.ListenAndServe(":"+port, handlers.CORS(originsOk, headersOk, methodsOk)(router)))
}
