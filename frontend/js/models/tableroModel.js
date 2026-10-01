// Modelo de Tablero - Estado en vivo de los llamados + flujo SSE
import { CONFIG, urlAPI, urlEventos } from '../config.js';

/** Eventos que publica el backend en el flujo de texto */
export const TIPOS_EVENTO = [
    'snapshot',
    'turno_creado',
    'turno_llamado',
    'turno_atendido',
    'turno_cancelado',
    'turno_vencido',
    'turno_actualizado',
    'servicio_cambiado'
];

/**
 * Consulta puntual del estado del tablero (respaldo del SSE)
 * @returns {Promise<Object>}
 */
export const obtenerTablero = async () => {
    const response = await fetch(urlAPI('/api/tablero'), {
        method: 'GET',
        headers: { 'Content-Type': 'application/json' }
    });

    const payload = await response.json().catch(() => ({}));

    if (!response.ok) {
        throw new Error(payload.error || 'No fue posible obtener el tablero');
    }

    return payload.data;
};

/**
 * Mantiene una conexión en vivo con el servidor.
 * Usa Server-Sent Events y cae a polling si la conexión no es posible.
 *
 * @param {{alRecibir: (tipo: string, datos: any) => void, alEstado: (estado: string) => void}} manejadores
 * @returns {() => void} Función para cerrar la conexión
 */
export const conectarEventos = ({ alRecibir, alEstado }) => {
    let fuente = null;
    let temporizador = null;
    let intervalo = null;
    let fallos = 0;
    let detenido = false;

    const detenerTemporizador = () => {
        if (temporizador) {
            window.clearTimeout(temporizador);
            temporizador = null;
        }
    };

    const detenerPolling = () => {
        if (intervalo) {
            window.clearInterval(intervalo);
            intervalo = null;
        }
    };

    const consultarPorPolling = async () => {
        try {
            const tablero = await obtenerTablero();
            alEstado('degradado');
            alRecibir('snapshot', tablero);
        } catch {
            alEstado('error');
        }
    };

    const iniciarPolling = () => {
        if (intervalo || detenido) return;
        alEstado('conectando');
        consultarPorPolling();
        intervalo = window.setInterval(consultarPorPolling, CONFIG.POLLING_INTERVAL);
    };

    const conectar = () => {
        if (detenido) return;

        if (typeof window.EventSource === 'undefined') {
            iniciarPolling();
            return;
        }

        try {
            fuente = new EventSource(urlEventos());
        } catch {
            iniciarPolling();
            return;
        }

        fuente.onopen = () => {
            fallos = 0;
            detenerPolling();
            alEstado('conectado');
        };

        fuente.onerror = () => {
            if (detenido) return;
            fallos += 1;

            if (fuente) {
                fuente.close();
                fuente = null;
            }

            // Tras varios intentos se confirma el respaldo por polling
            if (fallos >= 2) {
                iniciarPolling();
                return;
            }

            alEstado('conectando');
            detenerTemporizador();
            temporizador = window.setTimeout(conectar, CONFIG.POLLING_INTERVAL / 3);
        };

        TIPOS_EVENTO.forEach((tipo) => {
            fuente.addEventListener(tipo, (evento) => {
                try {
                    alRecibir(tipo, JSON.parse(evento.data));
                } catch {
                    // Evento con payload ilegible: se ignora
                }
            });
        });
    };

    alEstado('conectando');
    conectar();

    return () => {
        detenido = true;
        detenerTemporizador();
        detenerPolling();
        if (fuente) {
            fuente.close();
            fuente = null;
        }
    };
};