// Vista de Turnos - Renderizado del lado del cliente
import { escapar, renderHTML } from '../utils/dom.js';
import { formatearEstado, formatearFechaHora, mensajeCola, nombrePaciente } from '../utils/format.js';

/**
 * Muestra un mensaje provisional en el selector de servicios
 * @param {string} mensaje
 */
export const poblarAvisoServicios = (mensaje) => {
    const selector = document.getElementById('crear-servicio');
    if (selector) {
        selector.innerHTML = `<option value="">${escapar(mensaje)}</option>`;
        selector.disabled = true;
    }
};

/**
 * Llena el selector de servicios del formulario
 * @param {Array} servicios
 */
export const poblarServicios = (servicios) => {
    const selector = document.getElementById('crear-servicio');
    if (!selector) return;

    const opciones = servicios.map((servicio) => {
        const detalle = servicio.descripcion ? ` · ${servicio.descripcion}` : '';
        return `<option value="${servicio.id}">${escapar(servicio.nombre)} (${escapar(servicio.codigo)})${escapar(detalle)}</option>`;
    }).join('');

    selector.innerHTML = `<option value="">Selecciona un servicio…</option>${opciones}`;
    selector.disabled = servicios.length === 0;
};

/**
 * Llena el selector de tipos de atención según el servicio elegido
 * @param {Array} tipos
 * @param {string} [placeholder]
 */
export const poblarTiposAtencion = (tipos, placeholder = 'Selecciona el tipo de atención…') => {
    const selector = document.getElementById('crear-tipo');
    if (!selector) return;

    if (!tipos || tipos.length === 0) {
        selector.innerHTML = '<option value="">Este servicio no tiene tipos configurados</option>';
        selector.disabled = true;
        return;
    }

    selector.innerHTML = `<option value="">${escapar(placeholder)}</option>${
        tipos
            .map((tipo) => `<option value="${tipo.id}">${escapar(tipo.nombre)}</option>`)
            .join('')
    }`;
    selector.disabled = false;
};

/**
 * Construye el HTML de un ticket de turno
 * @param {Object} ticket
 * @returns {string}
 */
const plantillaTicket = (ticket) => {
    const estado = formatearEstado(ticket.estado);
    const cola = mensajeCola(ticket);

    const bloqueCola = cola ? `
        <div class="queue">
            <div class="queue-figure">${cola.titulo.match(/\d+/)?.[0] ?? ''}</div>
            <div class="queue-text">
                <p class="queue-title">${escapar(cola.titulo)}</p>
                <p class="queue-desc">${escapar(cola.descripcion)}</p>
            </div>
        </div>` : '';

    const pie = ticket.estado === 'EN_ATENCION'
        ? 'Dirígete al consultorio asignado.'
        : 'Guarda tu número de turno para consultar en cualquier momento.';

    return `
        <article class="ticket">
            <header class="ticket-top">
                <span class="ticket-codigo">${escapar(ticket.numero_turno || '—')}</span>
                <span class="badge ${estado.clase}">${escapar(estado.etiqueta)}</span>
            </header>
            <div class="ticket-body">
                <div>
                    <p class="ticket-field-label">Paciente</p>
                    <p class="ticket-field-value">${escapar(nombrePaciente(ticket))}</p>
                </div>
                <div>
                    <p class="ticket-field-label">Servicio</p>
                    <p class="ticket-field-value">${escapar(ticket.servicio_nombre || '—')}</p>
                </div>
                <div>
                    <p class="ticket-field-label">Tipo de atención</p>
                    <p class="ticket-field-value">${escapar(ticket.tipo_atencion_nombre || 'General')}</p>
                </div>
                <div>
                    <p class="ticket-field-label">Emitido</p>
                    <p class="ticket-field-value">${escapar(formatearFechaHora(ticket.creado_en))}</p>
                </div>
                ${ticket.cliente_documento ? `
                <div>
                    <p class="ticket-field-label">Documento</p>
                    <p class="ticket-field-value">${escapar(ticket.cliente_documento)}</p>
                </div>` : ''}
                ${ticket.cliente_contacto ? `
                <div>
                    <p class="ticket-field-label">Contacto</p>
                    <p class="ticket-field-value">${escapar(ticket.cliente_contacto)}</p>
                </div>` : ''}
            </div>
            ${bloqueCola}
            <footer class="ticket-foot">
                <span>${escapar(pie)}</span>
                <span>${escapar(ticket.motivo_cancelacion ? `Motivo: ${ticket.motivo_cancelacion}` : `Vence ${formatearFechaHora(ticket.expira_en)}`)}</span>
            </footer>
        </article>
    `;
};

/**
 * Renderiza un ticket en un contenedor
 * @param {string} idContenedor
 * @param {Object} ticket
 */
export const renderTicket = (idContenedor, ticket) => {
    if (!ticket) return;
    renderHTML(idContenedor, plantillaTicket(ticket));
};

/**
 * Renderiza el mensaje cuando no hay turno activo
 * @param {string} idContenedor
 * @param {string} [mensaje]
 */
export const renderSinTurno = (idContenedor, mensaje = 'No hay ningún turno activo con esos datos.') => {
    renderHTML(idContenedor, `
        <div class="empty">
            <div class="empty-icon" aria-hidden="true">🗒️</div>
            <p class="empty-title">Sin turno activo</p>
            <p class="small mt-1">${escapar(mensaje)}</p>
        </div>
    `);
};

/**
 * Renderiza la lista de turnos cuando el contacto tiene varios
 * @param {string} idContenedor
 * @param {Array} turnos
 */
export const renderListaTurnos = (idContenedor, turnos) => {
    if (!turnos || turnos.length === 0) {
        renderSinTurno(idContenedor);
        return;
    }

    const items = turnos.map((turno) => {
        const estado = formatearEstado(turno.estado);
        return `
            <button class="btn btn-outline btn-block mb-2" type="button"
                    style="justify-content: space-between;" data-numero="${escapar(turno.numero_turno || '')}">
                <span class="mono">${escapar(turno.numero_turno || '—')}</span>
                <span class="small muted">${escapar(turno.servicio_nombre || 'Servicio')}</span>
                <span class="badge ${estado.clase}">${escapar(estado.etiqueta)}</span>
            </button>
        `;
    }).join('');

    renderHTML(idContenedor, `
        <p class="small muted mb-3">Este contacto tiene ${turnos.length} turnos activos:</p>
        ${items}
    `);
};

/**
 * Muestra un estado de carga en un contenedor
 * @param {string} idContenedor
 */
export const renderCargando = (idContenedor) => {
    renderHTML(idContenedor, '<div class="skeleton" style="height: 220px;"></div>');
};

/**
 * Lee los valores del formulario de creación
 * @returns {{nombre: string, documento: string, contacto: string, servicio_id: string, tipo_atencion_id: string}}
 */
export const obtenerDatosCrear = () => ({
    nombre: document.getElementById('crear-nombre')?.value.trim() ?? '',
    documento: document.getElementById('crear-documento')?.value.trim() ?? '',
    contacto: document.getElementById('crear-contacto')?.value.trim() ?? '',
    servicio_id: document.getElementById('crear-servicio')?.value ?? '',
    tipo_atencion_id: document.getElementById('crear-tipo')?.value ?? ''
});

/**
 * Lee el valor consultado
 * @returns {string}
 */
export const obtenerConsulta = () =>
    document.getElementById('consultar-query')?.value.trim() ?? '';

/**
 * Restablece el formulario de creación conservando el servicio elegido
 */
export const limpiarFormularioCrear = () => {
    ['crear-nombre', 'crear-documento', 'crear-contacto'].forEach((id) => {
        const campo = document.getElementById(id);
        if (campo) campo.value = '';
    });
};