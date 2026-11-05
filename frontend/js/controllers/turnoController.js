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
import {
    compactar, normalizarContacto, pareceTelefono, validarContacto, validarDocumento,
    validarNombre, validarServicio, validarTipoAtencion
} from '../utils/validar.js';
import { descargarTicketPdf } from '../utils/ticketPdf.js';

/** Tickets mostrados en pantalla, por número de turno (para el PDF) */
const tickets = new Map();
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
        this._initValidacionEnVivo();
        this._initDescargaPdf();
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
     * Valida todos los campos del formulario y los marca en pantalla
     * @returns {Array<[string, string]>} Campos con error
     */
    _validarFormulario(datos) {
        const tipoRequerido = !document.getElementById('crear-tipo')?.disabled;
        const errores = [
            ['crear-servicio', validarServicio(datos.servicio_id)],
            ['crear-tipo', validarTipoAtencion(datos.tipo_atencion_id, tipoRequerido)],
            ['crear-nombre', validarNombre(datos.nombre)],
            ['crear-documento', validarDocumento(datos.documento)],
            ['crear-contacto', validarContacto(datos.contacto)]
        ];

        errores.forEach(([campo, mensaje]) => marcarErrorCampo(campo, mensaje));
        return errores.filter(([, mensaje]) => mensaje);
    },

    /**
     * Validación en vivo: filtra caracteres no permitidos y avisa al salir del campo
     */
    _initValidacionEnVivo() {
        const reglas = {
            'crear-servicio': () => validarServicio(document.getElementById('crear-servicio').value),
            'crear-tipo': () => validarTipoAtencion(
                document.getElementById('crear-tipo').value,
                !document.getElementById('crear-tipo').disabled
            ),
            'crear-nombre': () => validarNombre(document.getElementById('crear-nombre').value),
            'crear-documento': () => validarDocumento(document.getElementById('crear-documento').value),
            'crear-contacto': () => validarContacto(document.getElementById('crear-contacto').value)
        };

        const filtros = {
            'crear-nombre': (v) => v.replace(/[^\p{L}\s'.-]/gu, ''),
            'crear-documento': (v) => v.replace(/\D/g, ''),
            'crear-contacto': (v) => v.replace(/[^\d\s+().-]/g, '')
        };

        Object.entries(reglas).forEach(([id, regla]) => {
            const campo = document.getElementById(id);
            if (!campo) return;

            campo.addEventListener('input', () => {
                if (filtros[id]) {
                    const limpio = filtros[id](campo.value);
                    if (limpio !== campo.value) campo.value = limpio;
                }
                if (campo.classList.contains('has-error')) marcarErrorCampo(id, regla());
            });
            campo.addEventListener('blur', () => {
                if (campo.value !== '' || campo.required) marcarErrorCampo(id, regla());
            });
            campo.addEventListener('change', () => marcarErrorCampo(id, regla()));
        });
    },

    /**
     * Pinta un ticket y lo guarda para poder descargar su PDF
     */
    _mostrarTicket(idContenedor, ticket) {
        tickets.set(ticket.numero_turno, ticket);
        renderTicket(idContenedor, ticket);
    },

    /**
     * Botón "Descargar PDF" de los tickets mostrados
     */
    _initDescargaPdf() {
        ['resultado-crear', 'resultado-consulta'].forEach((id) => {
            delegar(document.getElementById(id), 'click', 'button[data-pdf]', ({ objetivo }) => {
                const ticket = tickets.get(objetivo.dataset.pdf);
                if (ticket && !descargarTicketPdf(ticket)) {
                    mostrarAlerta('crear-alerts', 'error', 'No fue posible generar el PDF');
                }
            });
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

        const errores = this._validarFormulario(datos);
        if (errores.length > 0) {
            document.getElementById(errores[0][0])?.focus();
            return;
        }

        estadoCarga('btn-crear', true, 'Generando…');

        try {
            const ticket = await crearTurno({
                nombre: compactar(datos.nombre),
                documento: datos.documento.replace(/[\s.-]/g, ''),
                contacto: normalizarContacto(datos.contacto),
                servicio_id: Number(datos.servicio_id),
                tipo_atencion_id: datos.tipo_atencion_id ? Number(datos.tipo_atencion_id) : 0
            });

            limpiarFormularioCrear();
            poblarTiposAtencion([]);
            this._mostrarTicket('resultado-crear', ticket);
            mostrarAlerta('crear-alerts', 'success', `Turno ${ticket.numero_turno} generado. Se descargó el ticket en PDF.`);
            // Equivale a la impresión del ticket: se genera el PDF automáticamente
            descargarTicketPdf(ticket);
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
        let consulta = obtenerConsulta();
        limpiarAlertas('consulta-alerts');

        if (!consulta) {
            marcarErrorCampo('consultar-query', 'Ingresa tu número de turno o de contacto');
            return;
        }

        if (pareceTelefono(consulta)) {
            // Un teléfono se consulta tal como se guardó: solo dígitos
            const error = validarContacto(consulta);
            if (error) {
                marcarErrorCampo('consultar-query', error);
                return;
            }
            consulta = normalizarContacto(consulta);
        } else if (!/^[A-Za-z]{2,4}-?\d{1,5}$/.test(consulta)) {
            marcarErrorCampo('consultar-query', 'Usa el formato del ticket, por ejemplo MED-003');
            return;
        } else {
            consulta = consulta.toUpperCase();
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

            this._mostrarTicket('resultado-consulta', respuesta.turno);
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