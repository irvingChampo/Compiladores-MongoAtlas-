package routes

import (
	"net/http"

	"github.com/gorilla/mux"
	"mongoapi/handlers"
)

func SetupRoutes() *mux.Router {
	router := mux.NewRouter()

	router.HandleFunc("/api/analizar", handlers.AnalizarHandler).Methods("POST")
	router.HandleFunc("/api/ejecutar", handlers.ExecuteHandler).Methods("POST")



	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Servidor activo"))
	})

	return router
}
