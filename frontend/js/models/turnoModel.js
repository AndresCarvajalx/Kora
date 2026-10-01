// Modelo de Turnos - Peticiones del cliente al backend
import { urlAPI } from '../config.js';

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
 * Crea un nuevo turno para el cliente
 * @param {{contacto: string, nombre: string, documento: string, servicio_id: number, tipo_atencion_id: number}} datos
 * @returns {Promise<Object>} Ticket generado con número y posición
 */
export const crearTurno = async (datos) => {
    const response = await fetch(urlAPI('/api/turnos'), {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(datos)
    });

    return leerRespuesta(response);
};

/**
 * Consulta un turno por número de turno o por número de contacto
 * @param {string} consulta - 'CAR-014' o '3001234567'
 * @returns {Promise<{turno_activo: boolean, turno: Object|null, turnos: Object[], mensaje: string}>}
 */
export const consultarTurno = async (consulta) => {
    const response = await fetch(urlAPI('/api/turnos/consultar'), {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ query: consulta })
    });

    return leerRespuesta(response);
};

/**
 * Obtiene un turno por su número visible
 * @param {string} numero - Ej: CAR-014
 * @returns {Promise<Object>}
 */
export const obtenerTurnoPorNumero = async (numero) => {
    const response = await fetch(urlAPI(`/api/turnos/${encodeURIComponent(numero)}`), {
        method: 'GET',
        headers: { 'Content-Type': 'application/json' }
    });

    return leerRespuesta(response);
};