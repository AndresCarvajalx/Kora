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
		EnAtencion:    toPunteros(enAtencion),
		Historial:     toPunteros(historial),
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

	realtime.Default().Publicar(tipo, map[string]interface{}{
		"tablero": tablero,
		"detalle": datosExtra,
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

// toPunteros convierte un slice de valores en un slice de punteros
func toPunteros(items []model.Turno) []*model.Turno {
	resultado := make([]*model.Turno, 0, len(items))
	for i := range items {
		resultado = append(resultado, &items[i])
	}
	return resultado
}
