// Vista de Tablero - Renderizado de los llamados en tiempo real
import { escapar, renderHTML } from '../utils/dom.js';
import { formatearHora, nombrePaciente } from '../utils/format.js';

/**
 * Normaliza el estado del tablero sea cual sea el origen del dato
 * @param {Object} tablero
 * @returns {{en_atencion: Array, historial: Array, pendientes: number}}
 */
const normalizar = (tablero) => ({
    en_atencion: tablero?.en_atencion ?? [],
    historial: tablero?.historial ?? [],
    pendientes: tablero?.pendientes ?? 0
});

/**
 * Renderiza la llamada actual
 * @param {string} idContenedor
 * @param {Object|null} turno
 */
export const renderLlamadoActual = (idContenedor, turno) => {
    const contenedor = document.getElementById(idContenedor);
    if (!contenedor) return;

    if (!turno) {
        contenedor.innerHTML = `
            <div class="llamado-actual is-idle">
                <p class="llamado-label">Llamando a</p>
                <p class="llamado-codigo">Sin llamados activos</p>
                <p class="llamado-servicio">Los turnos aparecerán aquí en cuanto se llamen.</p>
            </div>
        `;
        return;
    }

    contenedor.innerHTML = `
        <div class="llamado-actual" role="status" aria-live="assertive">
            <p class="llamado-label">Llamando a</p>
            <p class="llamado-codigo">${escapar(turno.numero_turno || '—')}</p>
            <p class="llamado-nombre">${escapar(nombrePaciente(turno))}</p>
            <p class="llamado-servicio">
                ${escapar(turno.servicio_nombre || 'Servicio')}
                ${turno.tipo_atencion_nombre ? ` · ${escapar(turno.tipo_atencion_nombre)}` : ''}
            </p>
        </div>
    `;
};

/**
 * Renderiza las tarjetas de los turnos en atención
 * @param {string} idContenedor
 * @param {Array} turnos
 */
export const renderEnAtencion = (idContenedor, turnos) => {
    const contenedor = document.getElementById(idContenedor);
    if (!contenedor) return;

    if (!turnos || turnos.length === 0) {
        contenedor.innerHTML = `
            <div class="empty">
                <div class="empty-icon" aria-hidden="true">🕗</div>
                <p class="empty-title">Nadie está siendo atendido ahora</p>
            </div>
        `;
        return;
    }

    contenedor.innerHTML = `<div class="llamados-grid">${turnos.map((turno) => `
        <article class="llamado-card">
            <span class="badge badge-servicio">${escapar(turno.servicio_codigo || '')}</span>
            <p class="llamado-card-codigo">${escapar(turno.numero_turno || '—')}</p>
            <p class="llamado-card-nombre">${escapar(nombrePaciente(turno))}</p>
            <p class="llamado-card-meta">
                ${escapar(turno.tipo_atencion_nombre || 'Atención general')}
                · ${escapar(formatearHora(turno.llamado_en))}
            </p>
        </article>
    `).join('')}</div>`;
};

/**
 * Renderiza el historial de últimos llamados
 * @param {string} idContenedor
 * @param {Array} turnos
 * @param {number} [limite]
 */
export const renderHistorial = (idContenedor, turnos, limite = 8) => {
    const contenedor = document.getElementById(idContenedor);
    if (!contenedor) return;

    const lista = (turnos || []).slice(0, limite);

    if (lista.length === 0) {
        contenedor.innerHTML = `
            <div class="empty">
                <p class="empty-title">Todavía no hay llamados registrados</p>
            </div>
        `;
        return;
    }

    contenedor.innerHTML = `<div class="ticker-list">${lista.map((turno) => `
        <div class="ticker-item">
            <div class="row" style="gap: var(--space-3);">
                <strong class="mono">${escapar(turno.numero_turno || '—')}</strong>
                <span>${escapar(nombrePaciente(turno))}</span>
                <span class="badge badge-neutro">${escapar(turno.servicio_nombre || '')}</span>
            </div>
            <time datetime="${escapar(turno.llamado_en || '')}">${escapar(formatearHora(turno.llamado_en))}</time>
        </div>
    `).join('')}</div>`;
};

/**
 * Panel compacto para la página del cliente
 * @param {string} idContenedor
 * @param {Object} tablero
 */
export const renderPanelCliente = (idContenedor, tablero) => {
    const { en_atencion, historial, pendientes } = normalizar(tablero);
    const principal = en_atencion[0] || null;

    renderHTML(idContenedor, `
        <div class="llamado-actual ${principal ? '' : 'is-idle'}">
            <p class="llamado-label">${principal ? 'Llamando ahora a' : 'Sin llamados activos'}</p>
            <p class="llamado-codigo">${principal ? escapar(principal.numero_turno || '—') : '—'}</p>
            ${principal ? `
                <p class="llamado-nombre">${escapar(nombrePaciente(principal))}</p>
                <p class="llamado-servicio">${escapar(principal.servicio_nombre || '')}</p>
            ` : '<p class="llamado-servicio">Los llamados aparecerán aquí automáticamente.</p>'}
        </div>

        <div class="row-between mt-4 mb-3">
            <h3>Últimos llamados</h3>
            <span class="badge badge-neutro">${pendientes} en espera</span>
        </div>
        <div id="panel-historial"></div>
    `);

    const destino = document.getElementById('panel-historial');
    if (destino) {
        renderHistorialEn(destino, historial.slice(0, 3));
    }
};

/**
 * Escribe el historial directamente sobre un nodo
 * @param {Element} nodo
 * @param {Array} turnos
 */
const renderHistorialEn = (nodo, turnos) => {
    if (!turnos || turnos.length === 0) {
        nodo.innerHTML = '<p class="small muted">Todavía no hay llamados registrados.</p>';
        return;
    }

    nodo.innerHTML = `<div class="ticker-list">${turnos.map((turno) => `
        <div class="ticker-item">
            <div class="row" style="gap: var(--space-3);">
                <strong class="mono">${escapar(turno.numero_turno || '—')}</strong>
                <span>${escapar(nombrePaciente(turno))}</span>
                <span class="badge badge-neutro">${escapar(turno.servicio_nombre || '')}</span>
            </div>
            <time>${escapar(formatearHora(turno.llamado_en))}</time>
        </div>
    `).join('')}</div>`;
};

/**
 * Actualiza el indicador de conexión del flujo en vivo
 * @param {string} idIndicador
 * @param {string} estado - conectado | conectando | error
 */
export const setEstadoConexion = (idIndicador, estado) => {
    const indicador = document.getElementById(idIndicador);
    if (!indicador) return;

    const etiquetas = {
        conectado: 'En vivo',
        degradado: 'Actualizando',
        conectando: 'Conectando',
        error: 'Sin conexión'
    };

    indicador.dataset.estado = estado;
    indicador.textContent = etiquetas[estado] || '—';
};

/**
 * Actualiza el reloj de la pantalla de tablero
 * @param {string} idReloj
 */
export const actualizarReloj = (idReloj) => {
    const reloj = document.getElementById(idReloj);
    if (!reloj) return;

    reloj.textContent = new Date().toLocaleTimeString('es', {
        hour: '2-digit',
        minute: '2-digit'
    });
};

/**
 * Actualiza el contador de pendientes del tablero
 * @param {number} pendientes
 */
export const setContadorPendientes = (pendientes) => {
    const contador = document.getElementById('contador-pendientes');
    if (contador) {
        contador.textContent = `${pendientes} en espera`;
    }
};

/**
 * Actualiza el texto del conteo de turnos en atención
 * @param {number} total
 */
export const setConteoAtencion = (total) => {
    const conteo = document.getElementById('conteo-atencion');
    if (conteo) {
        conteo.textContent = total === 0
            ? 'Ninguno por el momento'
            : `${total} ${total === 1 ? 'turno' : 'turnos'} en curso`;
    }
};