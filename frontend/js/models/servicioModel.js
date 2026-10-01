// Modelo de Servicios - Catálogo de servicios médicos del centro
import { CONFIG, urlAPI } from '../config.js';

let cache = null;
let cacheFecha = 0;

/**
 * Consulta el catálogo de servicios con sus tipos de atención.
 * El resultado se cachea en memoria durante CONFIG.CACHE_SERVICIOS.
 * @param {boolean} [forzar] - Ignora la caché
 * @returns {Promise<Array>}
 */
export const listarServicios = async (forzar = false) => {
    const vigente = cache && (Date.now() - cacheFecha) < CONFIG.CACHE_SERVICIOS;

    if (vigente && !forzar) return cache;

    const response = await fetch(urlAPI('/api/servicios'), {
        method: 'GET',
        headers: { 'Content-Type': 'application/json' }
    });

    const payload = await response.json();

    if (!response.ok) {
        throw new Error(payload.error || 'No fue posible cargar los servicios');
    }

    cache = Array.isArray(payload.data) ? payload.data : [];
    cacheFecha = Date.now();

    return cache;
};

/**
 * Busca un servicio por su código (CAR, MED, …)
 * @param {string} codigo
 * @returns {Object|undefined}
 */
export const buscarServicioPorCodigo = (codigo) =>
    (cache || []).find((servicio) => servicio.codigo === codigo);

/**
 * Vacía la caché del catálogo
 */
export const limpiarCacheServicios = () => {
    cache = null;
    cacheFecha = 0;
};