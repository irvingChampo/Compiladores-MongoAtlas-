package main

import (
	"log"
	"mongoapi/config"
	"mongoapi/routes"
	"net/http"
	"os"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

func main() {
	// Conexión a MongoDB
	config.ConnectMongo()

	// Configurar el router
	router := routes.SetupRoutes()

	// Configurar CORS
	headersOk := handlers.AllowedHeaders([]string{"Content-Type", "X-Requested-With", "Authorization"})
	originsOk := handlers.AllowedOrigins([]string{"http://localhost:5173", "https://mongo-front.onrender.com"})
	methodsOk := handlers.AllowedMethods([]string{"GET", "POST", "OPTIONS"})
	credentialsOk := handlers.AllowCredentials()

	// Configuración del puerto
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080" // Puerto por defecto para pruebas locales
	}

	// Aplicar CORS globalmente
	corsRouter := mux.NewRouter()
	corsRouter.PathPrefix("/").Handler(handlers.CORS(originsOk, headersOk, methodsOk, credentialsOk)(router))

	// Iniciar el servidor
	log.Printf("Servidor iniciado en :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, corsRouter))
}