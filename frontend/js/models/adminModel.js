// Modelo Admin - Peticiones de gestión de turnos (requiere token)
import { urlAPI } from '../config.js';
import { authModel } from './authModel.js';

/**
 * Desenvuelve la respuesta de la API y normaliza los errores
 * @param {Response} response
 * @returns {Promise<any>}
 */
const leerRespuesta = async (response) => {
    const payload = await response.json().catch(() => ({}));

    if (!response.ok) {
        const error = new Error(payload.error || 'Ocurrió un error inesperado');
        error.status = response.status;
        throw error;
    }

    return payload.data;
};

/**
 * Lista turnos aplicando filtros del panel
 * @param {{estado?: string, servicio_id?: number|string, numero?: string}} [filtros]
 * @returns {Promise<Array>}
 */
export const listarTurnos = async (filtros = {}) => {
    const params = new URLSearchParams();

    if (filtros.estado) params.set('estado', filtros.estado);
    if (filtros.servicio_id) params.set('servicio_id', filtros.servicio_id);
    if (filtros.numero) params.set('numero', filtros.numero);
    if (filtros.hoy) params.set('hoy', 'true');

    const query = params.toString();
    const url = query ? `${urlAPI('/api/admin/turnos')}?${query}` : urlAPI('/api/admin/turnos');

    const response = await fetch(url, {
        method: 'GET',
        headers: authModel.getAuthHeaders()
    });

    return leerRespuesta(response);
};

/**
 * Llama a un turno: PENDIENTE → EN_ATENCION
 * @param {number} id
 * @returns {Promise<Object>}
 */
export const llamarTurno = async (id) => {
    const response = await fetch(urlAPI(`/api/admin/turnos/${id}/llamar`), {
        method: 'POST',
        headers: authModel.getAuthHeaders()
    });

    return leerRespuesta(response);
};

/**
 * Marca un turno como atendido
 * @param {number} id
 * @returns {Promise<Object>}
 */
export const atenderTurno = async (id) => {
    const response = await fetch(urlAPI(`/api/admin/turnos/${id}/atender`), {
        method: 'POST',
        headers: authModel.getAuthHeaders()
    });

    return leerRespuesta(response);
};

/**
 * Cancela un turno
 * @param {number} id
 * @param {string} [motivo]
 * @returns {Promise<Object>}
 */
export const cancelarTurno = async (id, motivo = '') => {
    const response = await fetch(urlAPI(`/api/admin/turnos/${id}/cancelar`), {
        method: 'POST',
        headers: authModel.getAuthHeaders(),
        body: JSON.stringify({ motivo })
    });

    return leerRespuesta(response);
};

/**
 * Cambia el servicio (y cola) de un turno pendiente
 * @param {number} id
 * @param {{servicio_id: number, tipo_atencion_id?: number, motivo?: string}} datos
 * @returns {Promise<Object>}
 */
export const cambiarServicio = async (id, datos) => {
    const response = await fetch(urlAPI(`/api/admin/turnos/${id}/servicio`), {
        method: 'PATCH',
        headers: authModel.getAuthHeaders(),
        body: JSON.stringify(datos)
    });

    return leerRespuesta(response);
};

/**
 * Marca como vencidos los turnos que ya expiraron
 * @returns {Promise<{turnos_vencidos: number}>}
 */
export const vencerTurnos = async () => {
    const response = await fetch(urlAPI('/api/admin/turnos/vencer'), {
        method: 'POST',
        headers: authModel.getAuthHeaders()
    });

    return leerRespuesta(response);
};