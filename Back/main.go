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
    router := mux.NewRouter()

    // Configuración de CORS
    originsOk := handlers.AllowedOrigins([]string{"http://localhost:5173", "https://mongo-front.onrender.com"})
    headersOk := handlers.AllowedHeaders([]string{"X-Requested-With", "Content-Type", "Authorization"})
    methodsOk := handlers.AllowedMethods([]string{"GET", "POST", "OPTIONS"})

    // Configuración de rutas
    routes.SetRoutes(router)

    // Conexión a MongoDB
    config.ConnectDB()

    // Configuración del puerto
    port := os.Getenv("PORT")
    if port == "" {
        port = "8080" // Puerto por defecto para pruebas locales
    }

    // Iniciar el servidor
    log.Printf("Servidor iniciado en el puerto %s", port)
    log.Fatal(http.ListenAndServe(":"+port, handlers.CORS(originsOk, headersOk, methodsOk)(router)))
}