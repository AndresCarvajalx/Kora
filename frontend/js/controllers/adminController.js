// Controlador de Personal - Panel administrativo
import { delegar, estadoCarga, limpiarAlertas, mostrarAlerta } from '../utils/dom.js';
import { authModel } from '../models/authModel.js';
import * as adminModel from '../models/adminModel.js';
import { listarServicios } from '../models/servicioModel.js';
import { conectarEventos } from '../models/tableroModel.js';
import {
    abrirModalServicio,
    cambiarVista,
    poblarFiltroServicios,
    renderStats,
    renderTablaTurnos,
    setEstadoActivo
} from '../views/adminView.js';
import { setEstadoConexion } from '../views/tableroView.js';

/** Estado local del panel */
const estado = {
    filtroEstado: '',
    filtroServicio: '',
    busqueda: '',
    servicios: [],
    turnos: [],
    desconectar: null,
    cargando: false
};

export const adminController = {
    /**
     * Punto de entrada del panel de personal
     */
    init() {
        this._initLogin();
        this._initLogout();
        this._initFiltros();
        this._initAcciones();
        this._cargarServicios();

        if (authModel.estaAutenticado()) {
            cambiarVista(true);
            this._iniciarSesionViva();
        } else {
            cambiarVista(false);
        }
    },

    /* ---------------------------------------------------------------- */
    /* Autenticación                                                     */
    /* ---------------------------------------------------------------- */

    _initLogin() {
        const formulario = document.getElementById('form-login');
        if (!formulario) return;

        formulario.addEventListener('submit', async (evento) => {
            evento.preventDefault();
            limpiarAlertas('login-alerts');

            const email = document.getElementById('login-email').value.trim();
            const password = document.getElementById('login-password').value;

            if (!email || !password) {
                mostrarAlerta('login-alerts', 'error', 'Completa email y contraseña');
                return;
            }

            estadoCarga('btn-login', true, 'Ingresando…');

            try {
                await authModel.login(email, password);
                cambiarVista(true);
                await this.cargarTurnos();
                this._iniciarSesionViva();
                mostrarAlerta('dashboard-alerts', 'success', 'Sesión iniciada correctamente');
            } catch (error) {
                mostrarAlerta('login-alerts', 'error', error.message);
            } finally {
                estadoCarga('btn-login', false);
            }
        });
    },

    _initLogout() {
        document.getElementById('btn-logout')?.addEventListener('click', () => this.cerrarSesion());
    },

    /**
     * Cierra la sesión y detiene el flujo en vivo
     */
    cerrarSesion() {
        authModel.logout();
        if (estado.desconectar) {
            estado.desconectar();
            estado.desconectar = null;
        }
        cambiarVista(false);
    },

    /**
     * Activa la actualización en vivo del panel
     */
    _iniciarSesionViva() {
        if (estado.desconectar) return;

        estado.desconectar = conectarEventos({
            alEstado: (valor) => setEstadoConexion('estado-conexion', valor),
            alRecibir: (tipo) => {
                if (tipo === 'snapshot') return;
                if (!estado.cargando) this.cargarTurnos();
            }
        });
    },

    /* ---------------------------------------------------------------- */
    /* Catálogo y filtros                                                */
    /* ---------------------------------------------------------------- */

    async _cargarServicios() {
        try {
            estado.servicios = await listarServicios();
            poblarFiltroServicios(estado.servicios);
        } catch (error) {
            mostrarAlerta('dashboard-alerts', 'error', error.message);
        }
    },

    _initFiltros() {
        delegar(document.getElementById('filtros-estado'), 'click', '.chip', ({ objetivo }) => {
            estado.filtroEstado = objetivo.dataset.estado ?? '';
            setEstadoActivo(estado.filtroEstado);
            this.cargarTurnos();
        });

        document.getElementById('filtro-servicio')?.addEventListener('change', (evento) => {
            estado.filtroServicio = evento.target.value;
            this.cargarTurnos();
        });

        document.getElementById('filtro-busqueda')?.addEventListener('input', (evento) => {
            estado.busqueda = evento.target.value.trim().toLowerCase();
            this._pintar();
        });

        document.getElementById('btn-refrescar')?.addEventListener('click', () => this.cargarTurnos());
    },

    /* ---------------------------------------------------------------- */
    /* Datos                                                            */
    /* ---------------------------------------------------------------- */

    /**
     * Consulta los turnos del backend y los pinta
     */
    async cargarTurnos() {
        estado.cargando = true;

        try {
            estado.turnos = await adminModel.listarTurnos({
                estado: estado.filtroEstado,
                servicio_id: estado.filtroServicio
            });
            this._pintar();
        } catch (error) {
            if (error.status === 401) {
                this.cerrarSesion();
                mostrarAlerta('login-alerts', 'error', 'Tu sesión expiró, inicia sesión nuevamente');
                return;
            }
            mostrarAlerta('dashboard-alerts', 'error', error.message);
        } finally {
            estado.cargando = false;
        }
    },

    /**
     * Filtra en memoria por texto y renderiza tabla y resumen
     */
    _pintar() {
        const visibles = estado.busqueda
            ? estado.turnos.filter((turno) => {
                const texto = [
                    turno.numero_turno,
                    turno.cliente_nombre,
                    turno.cliente_contacto,
                    turno.cliente_documento,
                    turno.servicio_nombre
                ].filter(Boolean).join(' ').toLowerCase();
                return texto.includes(estado.busqueda);
            })
            : estado.turnos;

        renderTablaTurnos('tabla-turnos', visibles);
        renderStats('stats-container', estado.turnos);
    },

    /* ---------------------------------------------------------------- */
    /* Acciones sobre los turnos                                         */
    /* ---------------------------------------------------------------- */

    _initAcciones() {
        delegar(document.getElementById('tabla-turnos'), 'click', 'button[data-accion]', ({ objetivo }) => {
            const id = Number(objetivo.dataset.id);
            const accion = objetivo.dataset.accion;

            const acciones = {
                llamar: () => this.llamar(id),
                atender: () => this.atender(id),
                cancelar: () => this.cancelar(id),
                cambiar: () => this.cambiarServicio(id)
            };

            acciones[accion]?.();
        });
    },

    /**
     * Pasa un turno a EN_ATENCION
     * @param {number} id
     */
    async llamar(id) {
        await this._ejecutar(
            () => adminModel.llamarTurno(id),
            'Turno llamado, ve a la sala de atención'
        );
    },

    /**
     * Marca un turno como atendido
     * @param {number} id
     */
    async atender(id) {
        await this._ejecutar(
            () => adminModel.atenderTurno(id),
            'Turno marcado como atendido'
        );
    },

    /**
     * Cancela un turno pidiendo el motivo
     * @param {number} id
     */
    async cancelar(id) {
        const motivo = window.prompt('Motivo de la cancelación (opcional):', '');
        if (motivo === null) return;

        await this._ejecutar(
            () => adminModel.cancelarTurno(id, motivo.trim()),
            'Turno cancelado'
        );
    },

    /**
     * Re-encola un turno en el servicio elegido
     * @param {number} id
     */
    async cambiarServicio(id) {
        const turno = estado.turnos.find((item) => item.id === id);
        if (!turno) return;

        const datos = await abrirModalServicio(turno, estado.servicios);
        if (!datos) return;

        await this._ejecutar(
            () => adminModel.cambiarServicio(id, datos),
            'Turno movido de servicio y re-encolado'
        );
    },

    /**
     * Ejecuta una acción, avisa el resultado y refresca la lista
     * @param {() => Promise<any>} accion
     * @param {string} mensajeExito
     */
    async _ejecutar(accion, mensajeExito) {
        limpiarAlertas('dashboard-alerts');

        try {
            const turno = await accion();
            mostrarAlerta('dashboard-alerts', 'success', mensajeExito);
            await this.cargarTurnos();
            return turno;
        } catch (error) {
            if (error.status === 401) {
                this.cerrarSesion();
                mostrarAlerta('login-alerts', 'error', 'Tu sesión expiró, inicia sesión nuevamente');
                return null;
            }
            mostrarAlerta('dashboard-alerts', 'error', error.message);
            return null;
        }
    }
};