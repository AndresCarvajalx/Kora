package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"kora/backend/internal/database"
	"kora/backend/internal/handler"
	"kora/backend/internal/service"
)

func main() {
	ctx := context.Background()

	// Inicializar base de datos
	if err := database.InitDB(); err != nil {
		log.Fatalf("Error inicializando base de datos: %v", err)
	}
	defer database.CloseDB()

	log.Println("Conexión a base de datos establecida")

	// Aplicar migraciones pendientes
	if err := database.Migrate(ctx); err != nil {
		log.Fatalf("Error aplicando migraciones: %v", err)
	}
	log.Println("Migraciones aplicadas")

	// Limpieza periódica de turnos vencidos
	go limpiarTurnosVencidos()

	// Configurar rutas
	mux := http.NewServeMux()

	// Catálogo público
	mux.HandleFunc("/api/servicios", handler.ListarServiciosHandler)

	// Rutas del cliente (sin autenticación)
	mux.HandleFunc("/api/turnos/consultar", handler.ConsultarTurnoHandler)
	mux.HandleFunc("/api/turnos", handler.CrearTurnoHandler)
	mux.HandleFunc("/api/turnos/", handler.ObtenerTurnoHandler)

	// Tablero de llamados en tiempo real
	mux.HandleFunc("/api/tablero", handler.TableroHandler)
	mux.HandleFunc("/api/eventos", handler.EventosHandler)

	// Rutas del admin (con autenticación)
	mux.HandleFunc("/api/admin/login", handler.LoginHandler)
	mux.HandleFunc("/api/admin/turnos/pendientes", handler.AuthMiddleware(handler.ObtenerTurnosPendientesHandler))
	mux.HandleFunc("/api/admin/turnos/vencer", handler.AuthMiddleware(handler.MarcarTurnosVencidosHandler))
	mux.HandleFunc("/api/admin/turnos", handler.AuthMiddleware(handler.ListarTurnosHandler))
	mux.HandleFunc("/api/admin/turnos/", handler.AuthMiddleware(handler.RouterAdminTurnos))

	// Aplicar CORS
	handlerWithCORS := handler.CORS(mux)

	// Obtener puerto
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	srv := &http.Server{
		Addr:              fmt.Sprintf(":%s", port),
		Handler:           handlerWithCORS,
		ReadHeaderTimeout: 10 * time.Second,
		// Sin WriteTimeout: los flujos SSE son conexiones largas.
	}

	go func() {
		log.Printf("Servidor escuchando en %s", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Error levantando servidor: %v", err)
		}
	}()

	// Apagado ordenado
	parar := make(chan os.Signal, 1)
	signal.Notify(parar, os.Interrupt, syscall.SIGTERM)
	<-parar

	log.Println("Cerrando servidor...")
	contexto, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(contexto); err != nil {
		log.Printf("Error cerrando servidor: %v", err)
	}
}

// limpiarTurnosVencidos marca como vencidos los turnos cuya expiración pasó y
// avisa a los paneles conectados.
func limpiarTurnosVencidos() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		count, err := service.MarcarTurnosVencidos(ctx)
		cancel()

		if err != nil {
			log.Printf("Error marcando turnos vencidos: %v", err)
			continue
		}

		if count > 0 {
			log.Printf("Turnos marcados como vencidos: %d", count)
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			service.PublicarSnapshot(ctx)
			cancel()
		}
	}
}
