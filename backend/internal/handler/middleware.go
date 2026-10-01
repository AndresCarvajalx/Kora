package handler

import (
	"encoding/json"
	"log"
	"net/http"
)

// ErrorInterno registra el error real en el log y responde un mensaje genérico
func ErrorInterno(w http.ResponseWriter, err error) {
	log.Printf("error interno: %v", err)
	ErrorResponse(w, http.StatusInternalServerError, "error interno del servidor")
}

// CORS middleware habilita CORS para todas las rutas
func CORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// JSONResponse escribe una respuesta JSON con el código de estado dado
func JSONResponse(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}

// SuccessResponse responde con {"data": ...}
func SuccessResponse(w http.ResponseWriter, statusCode int, data interface{}) {
	JSONResponse(w, statusCode, map[string]interface{}{
		"data": data,
	})
}

// ErrorResponse responde con {"error": "..."}
func ErrorResponse(w http.ResponseWriter, statusCode int, message string) {
	JSONResponse(w, statusCode, map[string]string{
		"error": message,
	})
}
