// Vista de Personal - Renderizado del panel administrativo
import { $, delegar, escapar, renderHTML } from '../utils/dom.js';
import { formatearEstado, formatearHora, nombrePaciente } from '../utils/format.js';

/**
 * Muestra el resumen de turnos por estado
 * @param {string} idContenedor
 * @param {Array} turnos
 */
export const renderStats = (idContenedor, turnos) => {
    const lista = turnos || [];
    const contar = (estado) => lista.filter((turno) => turno.estado === estado).length;

    const tarjetas = [
        { clase: 'stat-pendiente', valor: contar('PENDIENTE'), etiqueta: 'Pendientes' },
        { clase: 'stat-atencion', valor: contar('EN_ATENCION'), etiqueta: 'En atención' },
        { clase: 'stat-atendido', valor: contar('ATENDIDO'), etiqueta: 'Atendidos' },
        { clase: 'stat-total', valor: lista.length, etiqueta: 'Total' }
    ];

    renderHTML(idContenedor, tarjetas.map((tarjeta) => `
        <article class="stat ${tarjeta.clase}">
            <p class="stat-value">${tarjeta.valor}</p>
            <p class="stat-label">${escapar(tarjeta.etiqueta)}</p>
        </article>
    `).join(''));
};

/**
 * Botones disponibles según el estado del turno
 * @param {Object} turno
 * @returns {string}
 */
const accionesDe = (turno) => {
    if (turno.estado === 'PENDIENTE') {
        return `
            <button class="btn btn-primary btn-sm" type="button" data-accion="llamar" data-id="${turno.id}">Llamar</button>
            <button class="btn btn-outline btn-sm" type="button" data-accion="cambiar" data-id="${turno.id}">Cambiar servicio</button>
            <button class="btn btn-danger btn-sm" type="button" data-accion="cancelar" data-id="${turno.id}">Cancelar</button>
        `;
    }

    if (turno.estado === 'EN_ATENCION') {
        return `
            <button class="btn btn-success btn-sm" type="button" data-accion="atender" data-id="${turno.id}">Atender</button>
            <button class="btn btn-danger btn-sm" type="button" data-accion="cancelar" data-id="${turno.id}">Cancelar</button>
        `;
    }

    return '<span class="small muted">—</span>';
};

/**
 * Renderiza la tabla de turnos
 * @param {string} idContenedor
 * @param {Array} turnos
 */
export const renderTablaTurnos = (idContenedor, turnos) => {
    const contenedor = document.getElementById(idContenedor);
    if (!contenedor) return;

    if (!turnos || turnos.length === 0) {
        contenedor.innerHTML = `
            <div class="empty">
                <div class="empty-icon" aria-hidden="true">📋</div>
                <p class="empty-title">No hay turnos que coincidan</p>
                <p class="small mt-1">Prueba con otro filtro o espera a que se generen nuevos turnos.</p>
            </div>
        `;
        return;
    }

    const filas = turnos.map((turno) => {
        const estado = formatearEstado(turno.estado);

        return `
            <tr data-id="${turno.id}">
                <td>
                    <span class="cell-codigo">${escapar(turno.numero_turno || '—')}</span>
                    <div class="cell-muted">${escapar(turno.servicio_nombre || 'Sin servicio')}</div>
                </td>
                <td>
                    <span class="strong">${escapar(nombrePaciente(turno))}</span>
                    <div class="cell-muted">${escapar(turno.tipo_atencion_nombre || 'Atención general')}</div>
                </td>
                <td class="cell-muted">
                    ${escapar(turno.cliente_contacto)}
                    ${turno.cliente_documento ? `<div>Doc. ${escapar(turno.cliente_documento)}</div>` : ''}
                </td>
                <td><span class="badge ${estado.clase}">${escapar(estado.etiqueta)}</span></td>
                <td class="cell-muted">${escapar(formatearHora(turno.llamado_en || turno.creado_en))}</td>
                <td class="cell-actions">${accionesDe(turno)}</td>
            </tr>
        `;
    }).join('');

    contenedor.innerHTML = `
        <div class="table-wrap">
            <table class="table">
                <caption class="sr-only">Listado de turnos del centro</caption>
                <thead>
                    <tr>
                        <th scope="col">Turno</th>
                        <th scope="col">Paciente</th>
                        <th scope="col">Contacto</th>
                        <th scope="col">Estado</th>
                        <th scope="col">Hora</th>
                        <th scope="col"><span class="sr-only">Acciones</span></th>
                    </tr>
                </thead>
                <tbody>${filas}</tbody>
            </table>
        </div>
    `;
};

/**
 * Llena el filtro de servicios
 * @param {Array} servicios
 * @param {string} [idSelector]
 */
export const poblarFiltroServicios = (servicios, idSelector = 'filtro-servicio') => {
    const selector = document.getElementById(idSelector);
    if (!selector) return;

    const actual = selector.value;
    selector.innerHTML = `<option value="">Todos los servicios</option>${
        servicios.map((servicio) => `<option value="${servicio.id}">${escapar(servicio.nombre)}</option>`).join('')
    }`;
    selector.value = actual;
};

/**
 * Marca el chip de estado activo
 * @param {string} estado
 */
export const setEstadoActivo = (estado) => {
    document.querySelectorAll('#filtros-estado .chip').forEach((chip) => {
        chip.classList.toggle('is-active', chip.dataset.estado === estado);
    });
};

/**
 * Abre el modal para cambiar el servicio de un turno
 * @param {Object} turno
 * @param {Array} servicios
 * @returns {{servicio_id: string, tipo_atencion_id: string, motivo: string}|null}
 */
export const abrirModalServicio = (turno, servicios) => {
    const raiz = document.getElementById('modal-root');
    if (!raiz) return null;

    const disponibles = servicios.filter((servicio) => servicio.id !== turno.servicio_id);

    raiz.innerHTML = `
        <div class="overlay" role="dialog" aria-modal="true" aria-labelledby="titulo-modal">
            <div class="modal">
                <div class="card-head">
                    <div>
                        <h2 class="card-title" id="titulo-modal">Cambiar de servicio</h2>
                        <p class="card-sub">
                            Turno <strong class="mono">${escapar(turno.numero_turno || '')}</strong> ·
                            ${escapar(nombrePaciente(turno))}
                        </p>
                    </div>
                </div>

                <form id="form-modal-servicio">
                    <div class="field">
                        <label class="field-label" for="modal-servicio">Nuevo servicio <span class="req">*</span></label>
                        <select class="select" id="modal-servicio" required>
                            <option value="">Selecciona un servicio…</option>
                            ${disponibles.map((servicio) => `<option value="${servicio.id}">${escapar(servicio.nombre)}</option>`).join('')}
                        </select>
                    </div>

                    <div class="field">
                        <label class="field-label" for="modal-tipo">Tipo de atención</label>
                        <select class="select" id="modal-tipo" disabled>
                            <option value="">Primero elige un servicio</option>
                        </select>
                    </div>

                    <div class="field">
                        <label class="field-label" for="modal-motivo">Motivo del cambio</label>
                        <input class="input" type="text" id="modal-motivo" placeholder="Ej: reasignado por el personal" maxlength="255">
                    </div>

                    <div class="modal-actions">
                        <button class="btn btn-ghost" type="button" data-accion="cerrar-modal">Cancelar</button>
                        <button class="btn btn-primary" type="submit">Cambiar servicio</button>
                    </div>
                </form>
            </div>
        </div>
    `;

    const overlay = raiz.querySelector('.overlay');

    return new Promise((resolver) => {
        const cerrar = (resultado) => {
            raiz.innerHTML = '';
            document.removeEventListener('keydown', alPulsarEscape);
            resolver(resultado);
        };

        const alPulsarEscape = (evento) => {
            if (evento.key === 'Escape') cerrar(null);
        };

        document.addEventListener('keydown', alPulsarEscape);

        overlay.addEventListener('click', (evento) => {
            if (evento.target === overlay) cerrar(null);
        });

        delegar(raiz, 'click', '[data-accion="cerrar-modal"]', () => cerrar(null));

        const selectorServicio = $('#modal-servicio', raiz);
        const selectorTipo = $('#modal-tipo', raiz);

        selectorServicio.addEventListener('change', () => {
            const servicio = servicios.find((item) => String(item.id) === selectorServicio.value);
            const tipos = servicio?.tipos_atencion ?? [];

            if (tipos.length === 0) {
                selectorTipo.innerHTML = '<option value="">Este servicio no tiene tipos configurados</option>';
                selectorTipo.disabled = true;
                return;
            }

            selectorTipo.innerHTML = '<option value="">Selecciona el tipo de atención…</option>' +
                tipos.map((tipo) => `<option value="${tipo.id}">${escapar(tipo.nombre)}</option>`).join('');
            selectorTipo.disabled = false;
        });

        $('#form-modal-servicio', raiz).addEventListener('submit', (evento) => {
            evento.preventDefault();

            const servicioId = selectorServicio.value;
            if (!servicioId) {
                selectorServicio.classList.add('has-error');
                return;
            }

            cerrar({
                servicio_id: Number(servicioId),
                tipo_atencion_id: selectorTipo.value ? Number(selectorTipo.value) : 0,
                motivo: $('#modal-motivo', raiz).value.trim()
            });
        });
    });
};

/**
 * Alterna entre login y panel
 * @param {boolean} mostrarPanel
 */
export const cambiarVista = (mostrarPanel) => {
    document.getElementById('login-section')?.classList.toggle('hidden', mostrarPanel);
    document.getElementById('dashboard-section')?.classList.toggle('hidden', !mostrarPanel);
};