// Utilidades de formato y presentación

/** Etiquetas y clases de cada estado de turno */
export const ESTADOS = {
    PENDIENTE: { etiqueta: 'Pendiente', clase: 'badge-pendiente' },
    EN_ATENCION: { etiqueta: 'En atención', clase: 'badge-en-atencion' },
    ATENDIDO: { etiqueta: 'Atendido', clase: 'badge-atendido' },
    CANCELADO: { etiqueta: 'Cancelado', clase: 'badge-cancelado' },
    VENCIDO: { etiqueta: 'Vencido', clase: 'badge-vencido' }
};

/**
 * Devuelve la etiqueta y clase CSS de un estado
 * @param {string} estado
 * @returns {{etiqueta: string, clase: string}}
 */
export const formatearEstado = (estado) =>
    ESTADOS[estado] || { etiqueta: estado || '—', clase: 'badge-neutro' };

/**
 * Formatea una fecha ISO como fecha y hora local
 * @param {string} iso
 * @returns {string}
 */
export const formatearFechaHora = (iso) => {
    if (!iso) return '—';
    const fecha = new Date(iso);
    if (Number.isNaN(fecha.getTime())) return '—';
    return fecha.toLocaleString('es', { dateStyle: 'medium', timeStyle: 'short' });
};

/**
 * Formatea una fecha ISO solo con la hora
 * @param {string} iso
 * @returns {string}
 */
export const formatearHora = (iso) => {
    if (!iso) return '—';
    const fecha = new Date(iso);
    if (Number.isNaN(fecha.getTime())) return '—';
    return fecha.toLocaleTimeString('es', { hour: '2-digit', minute: '2-digit' });
};

/**
 * Convierte minutos de espera a un texto legible
 * @param {number} minutos
 * @returns {string}
 */
export const formatearEspera = (minutos) => {
    const valor = Number(minutos);
    if (!Number.isFinite(valor) || valor <= 0) return 'Menos de 5 min';

    if (valor < 60) return `~${Math.round(valor)} min`;

    const horas = Math.floor(valor / 60);
    const resto = Math.round(valor % 60);
    return resto === 0 ? `~${horas} h` : `~${horas} h ${resto} min`;
};

/**
 * Muestra el nombre del paciente o su contacto si no tiene nombre
 * @param {Object} turno
 * @returns {string}
 */
export const nombrePaciente = (turno) =>
    (turno?.cliente_nombre || '').trim() || turno?.cliente_contacto || 'Paciente';

/**
 * Devuelve el texto de cola según el estado del turno
 * @param {Object} ticket
 * @returns {{titulo: string, descripcion: string}|null}
 */
export const mensajeCola = (ticket) => {
    if (!ticket || ticket.estado !== 'PENDIENTE') return null;

    const antes = ticket.personas_antes ?? 0;
    const posicion = ticket.posicion ?? antes + 1;
    const total = ticket.total_en_cola ?? posicion;

    return {
        titulo: `Posición ${posicion} de ${total} en la fila`,
        descripcion: antes === 0
            ? 'Eres el siguiente en ser llamado.'
            : `Hay ${antes} ${antes === 1 ? 'persona' : 'personas'} antes de ti · ${formatearEspera(ticket.tiempo_estimado_min)} estimado.`
    };
};