// Controlador de Tablero - Pantalla de llamados en tiempo real
import { renderHTML } from '../utils/dom.js';
import { activarSonido, reproducirAviso, sonidoActivo } from '../utils/audio.js';
import { conectarEventos, obtenerTablero } from '../models/tableroModel.js';
import {
    renderEnAtencion,
    renderHistorial,
    renderLlamadoActual,
    setContadorPendientes,
    setConteoAtencion,
    setEstadoConexion,
    actualizarReloj
} from '../views/tableroView.js';

/** Número del último turnoShown, para detectar llamados nuevos */
let ultimoNumero = null;
let desconectar = null;
let reloj = null;

export const tableroController = {
    /**
     * Inicializa la pantalla del tablero
     */
    init() {
        this._initReloj();
        this._initSonido();
        this._conectar();

        renderHTML('llamado-actual', '<div class="skeleton" style="height: 200px;"></div>');
        renderHTML('lista-atencion', '<div class="skeleton" style="height: 120px;"></div>');
    },

    /**
     * Reloj de la sala de espera
     */
    _initReloj() {
        actualizarReloj('reloj');
        reloj = window.setInterval(() => actualizarReloj('reloj'), 15000);
    },

    /**
     * Botón de aviso sonoro
     */
    _initSonido() {
        const boton = document.getElementById('btn-sonido');
        if (!boton) return;

        boton.addEventListener('click', () => {
            const activo = activarSonido(!sonidoActivo());
            boton.textContent = activo ? 'Silenciar' : 'Activar sonido';
            boton.setAttribute('aria-pressed', String(activo));
        });
    },

    /**
     * Conexión con el flujo de eventos
     */
    _conectar() {
        const pintar = (tablero) => this._pintar(tablero, { avisar: true });

        obtenerTablero()
            .then((tablero) => this._pintar(tablero, { avisar: false }))
            .catch(() => setEstadoConexion('estado-conexion', 'error'));

        desconectar = conectarEventos({
            alEstado: (estado) => setEstadoConexion('estado-conexion', estado),
            alRecibir: (tipo, datos) => {
                const tablero = datos?.tablero ?? (tipo === 'snapshot' ? datos : null);

                if (tablero) {
                    pintar(tablero);
                    return;
                }

                obtenerTablero()
                    .then((estado) => this._pintar(estado, { avisar: true }))
                    .catch(() => setEstadoConexion('estado-conexion', 'error'));
            }
        });
    },

    /**
     * Dibuja el tablero completo
     * @param {Object} tablero
     * @param {{avisar: boolean}} opciones
     */
    _pintar(tablero, { avisar }) {
        if (!tablero) return;

        const enAtencion = tablero.en_atencion ?? [];
        const historial = tablero.historial ?? [];
        const principal = enAtencion[0] || null;

        renderLlamadoActual('llamado-actual', principal);
        renderEnAtencion('lista-atencion', enAtencion);
        renderHistorial('historial-llamados', historial);
        setContadorPendientes(tablero.pendientes ?? 0);
        setConteoAtencion(enAtencion.length);

        // Aviso sonoro cuando cambia el turno llamado
        const numeroActual = principal?.numero_turno ?? null;
        if (avisar && numeroActual && numeroActual !== ultimoNumero) {
            reproducirAviso();
        }
        ultimoNumero = numeroActual;
    },

    /**
     * Limpia temporizadores y conexión
     */
    destroy() {
        if (desconectar) desconectar();
        if (reloj) window.clearInterval(reloj);
    }
};