// Controlador de Turnos - Página del paciente
import {
    estadoCarga, limpiarAlertas, marcarErrorCampo, mostrarAlerta, delegar, renderHTML
} from '../utils/dom.js';
import { crearTurno, consultarTurno } from '../models/turnoModel.js';
import { listarServicios } from '../models/servicioModel.js';
import { conectarEventos, obtenerTablero } from '../models/tableroModel.js';
import {
    limpiarFormularioCrear,
    obtenerConsulta,
    obtenerDatosCrear,
    poblarAvisoServicios,
    poblarServicios,
    poblarTiposAtencion,
    renderCargando,
    renderListaTurnos,
    renderSinTurno,
    renderTicket
} from '../views/turnoView.js';
import { renderPanelCliente, setEstadoConexion } from '../views/tableroView.js';

/** Contenido del catálogo de servicios en memoria */
let servicios = [];
/** Cierre de la conexión en vivo */
let desconectar = null;

export const turnoController = {
    /**
     * Inicializa la página del paciente
     */
    init() {
        this._initTabs();
        this._initFormCrear();
        this._initFormConsultar();
        this._cargarServicios();
        this._initTableroEnVivo();
    },

    /**
     * Alterna entre consulta por número de turno y por contacto
     */
    _initTabs() {
        const tabs = document.querySelectorAll('.tabs .tab');
        if (tabs.length === 0) return;

        const etiqueta = document.getElementById('label-consulta');
        const pista = document.getElementById('hint-consulta');
        const campo = document.getElementById('consultar-query');

        delegar(document.querySelector('.tabs'), 'click', '.tab', ({ objetivo }) => {
            const modo = objetivo.dataset.modo;

            tabs.forEach((tab) => {
                const activo = tab === objetivo;
                tab.classList.toggle('is-active', activo);
                tab.setAttribute('aria-selected', String(activo));
            });

            if (modo === 'numero') {
                etiqueta.textContent = 'Número de turno';
                pista.textContent = 'Tal como aparece en tu ticket, ej: CAR-014';
                campo.placeholder = 'Ej: CAR-014';
                campo.classList.add('mono');
            } else {
                etiqueta.textContent = 'Número de contacto';
                pista.textContent = 'El teléfono o correo con el que registraste tu turno';
                campo.placeholder = 'Ej: 300 123 4567';
                campo.classList.remove('mono');
            }
        });
    },

    /**
     * Carga el catálogo y arma el formulario de creación
     */
    async _cargarServicios() {
        poblarAvisoServicios('Cargando servicios…');

        try {
            servicios = await listarServicios();
            poblarServicios(servicios);
        } catch (error) {
            poblarAvisoServicios('No fue posible cargar los servicios');
            mostrarAlerta('crear-alerts', 'error', error.message);
            return;
        }

        document.getElementById('crear-servicio')?.addEventListener('change', (evento) => {
            const servicio = servicios.find((item) => String(item.id) === evento.target.value);
            poblarTiposAtencion(servicio?.tipos_atencion ?? []);
        });
    },

    /**
     * Registra el envío del formulario de creación
     */
    _initFormCrear() {
        const formulario = document.getElementById('form-crear');
        if (!formulario) return;

        formulario.addEventListener('submit', (evento) => {
            evento.preventDefault();
            this._crearTurno();
        });
    },

    /**
     * Registra el envío del formulario de consulta
     */
    _initFormConsultar() {
        const formulario = document.getElementById('form-consultar');
        if (!formulario) return;

        formulario.addEventListener('submit', (evento) => {
            evento.preventDefault();
            this._consultarTurno();
        });
    },

    /**
     * Valida y envía el turno nuevo
     */
    async _crearTurno() {
        const datos = obtenerDatosCrear();
        limpiarAlertas('crear-alerts');
        renderHTML('resultado-crear', '');

        const errores = [];
        if (!datos.servicio_id) errores.push(['crear-servicio', 'Selecciona un servicio médico']);
        if (datos.nombre.length < 3) errores.push(['crear-nombre', 'Ingresa tu nombre completo']);
        if (datos.contacto.length < 4) errores.push(['crear-contacto', 'Ingresa un número de contacto válido']);

        errores.forEach(([campo, mensaje]) => marcarErrorCampo(campo, mensaje));
        if (errores.length > 0) return;

        estadoCarga('btn-crear', true, 'Generando…');

        try {
            const ticket = await crearTurno({
                nombre: datos.nombre,
                documento: datos.documento,
                contacto: datos.contacto,
                servicio_id: Number(datos.servicio_id),
                tipo_atencion_id: datos.tipo_atencion_id ? Number(datos.tipo_atencion_id) : 0
            });

            limpiarFormularioCrear();
            poblarTiposAtencion([]);
            renderTicket('resultado-crear', ticket);
            mostrarAlerta('crear-alerts', 'success', `Turno ${ticket.numero_turno} generado. Guárdalo para consultar después.`);
        } catch (error) {
            mostrarAlerta('crear-alerts', 'error', error.message);
        } finally {
            estadoCarga('btn-crear', false);
        }
    },

    /**
     * Consulta el turno por número o por contacto
     */
    async _consultarTurno() {
        const consulta = obtenerConsulta();
        limpiarAlertas('consulta-alerts');

        if (!consulta) {
            marcarErrorCampo('consultar-query', 'Ingresa tu número de turno o de contacto');
            return;
        }

        marcarErrorCampo('consultar-query', '');
        estadoCarga('btn-consultar', true, 'Buscando…');
        renderCargando('resultado-consulta');

        try {
            const respuesta = await consultarTurno(consulta);

            if (!respuesta.turno_activo) {
                renderSinTurno('resultado-consulta', respuesta.mensaje || 'No hay turnos activos con esos datos.');
                return;
            }

            if (respuesta.turnos && respuesta.turnos.length > 1) {
                renderListaTurnos('resultado-consulta', respuesta.turnos);
                return;
            }

            renderTicket('resultado-consulta', respuesta.turno);
        } catch (error) {
            renderSinTurno('resultado-consulta', error.message);
            mostrarAlerta('consulta-alerts', 'error', error.message);
        } finally {
            estadoCarga('btn-consultar', false);
        }
    },

    /**
     * Conecta el panel de llamados al flujo en vivo
     */
    _initTableroEnVivo() {
        renderCargando('panel-llamados');

        const pintar = (tablero) => {
            if (tablero) renderPanelCliente('panel-llamados', tablero);
        };

        const recargar = () => {
            obtenerTablero().then(pintar).catch(() => setEstadoConexion('estado-conexion', 'error'));
        };

        obtenerTablero().then(pintar).catch(() => {
            renderHTML('panel-llamados', `
                <div class="empty">
                    <p class="empty-title">No fue posible cargar los llamados</p>
                    <p class="small mt-1">Reintenta en unos segundos.</p>
                </div>
            `);
        });

        desconectar = conectarEventos({
            alEstado: (estado) => setEstadoConexion('estado-conexion', estado),
            alRecibir: (tipo, datos) => {
                const tablero = datos?.tablero ?? (tipo === 'snapshot' ? datos : null);
                if (tablero) pintar(tablero);
                else recargar();
            }
        });

        // Permite elegir un turno desde la lista de turnos del contacto
        delegar(document.getElementById('resultado-consulta'), 'click', 'button[data-numero]', ({ objetivo }) => {
            const campo = document.getElementById('consultar-query');
            campo.value = objetivo.dataset.numero;
            this._consultarTurno();
        });
    },

    /**
     * Cierra la conexión en vivo
     */
    destroy() {
        if (desconectar) desconectar();
    }
};