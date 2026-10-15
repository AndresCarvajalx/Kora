package service

import (
	"context"
	"time"

	"kora/backend/internal/model"
	"kora/backend/internal/realtime"
	"kora/backend/internal/repository"
)

// ObtenerTablero arma el estado en vivo que consumen los paneles de llamados
func ObtenerTablero(ctx context.Context) (*model.Tablero, error) {
	enAtencion, err := repository.ListarEnAtencion(ctx)
	if err != nil {
		return nil, err
	}

	historial, err := repository.ListarUltimosLlamados(ctx, 12)
	if err != nil {
		return nil, err
	}

	pendientes, err := repository.ContarPendientes(ctx)
	if err != nil {
		return nil, err
	}

	tablero := &model.Tablero{
		ActualizadoEn: time.Now(),
		Pendientes:    pendientes,
		EnAtencion:    toTurnosTablero(enAtencion),
		Historial:     toTurnosTablero(historial),
	}

	if len(enAtencion) > 0 {
		tablero.UltimoServicio = enAtencion[0].ServicioNombre
	}

	return tablero, nil
}

// PublicarTablero notifica a los suscriptores con el estado actual del tablero
func PublicarTablero(ctx context.Context, tipo string, datosExtra interface{}) {
	tablero, err := ObtenerTablero(ctx)
	if err != nil {
		return
	}

	// Sanitizar datosExtra si es un *model.Turno para que SSE no filtre teléfono o documento
	detalle := datosExtra
	if t, ok := datosExtra.(*model.Turno); ok && t != nil {
		detalle = t.ATurnoTablero()
	}

	realtime.Default().Publicar(tipo, map[string]interface{}{
		"tablero": tablero,
		"detalle": detalle,
	})
}

// PublicarSnapshot envía el estado completo del tablero sin evento asociado
func PublicarSnapshot(ctx context.Context) {
	tablero, err := ObtenerTablero(ctx)
	if err != nil {
		return
	}

	realtime.Default().Publicar(realtime.EventoSnapshot, tablero)
}

// toTurnosTablero convierte un slice de turnos en un slice de punteros a TurnoTablero
// asegurando que teléfono y documento no se incluyan
func toTurnosTablero(items []model.Turno) []*model.TurnoTablero {
	resultado := make([]*model.TurnoTablero, 0, len(items))
	for i := range items {
		resultado = append(resultado, items[i].ATurnoTablero())
	}
	return resultado
}
