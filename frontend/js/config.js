// Configuración global del frontend

/**
 * Lee una clave de localStorage sin romper la aplicación si el
 * navegador bloquea el almacenamiento (modo privado, iframes, etc.).
 * @param {string} clave
 * @returns {string|null}
 */
const leerAjuste = (clave) => {
    try {
        return window.localStorage.getItem(clave);
    } catch {
        return null;
    }
};

/**
 * Resuelve la URL de la API en el entorno actual.
 * 1. Si existe una URL guardada en localStorage se respeta (configuración manual).
 * 2. En local el backend escucha en el puerto 8080.
 * 3. En cualquier otro host se asume que la API se sirve en el mismo origen.
 * @returns {string}
 */
const resolverAPI = () => {
    const guardada = leerAjuste('kora_api_url');
    if (guardada) return guardada.replace(/\/$/, '');

    const esLocal = ['localhost', '127.0.0.1', ''].includes(window.location.hostname);
    return esLocal ? 'http://localhost:8080' : window.location.origin;
};

export const CONFIG = {
    API_URL: resolverAPI(),
    APP_NAME: 'Kora',
    // Intervalo del respaldo por polling cuando el flujo SSE no está disponible
    POLLING_INTERVAL: 15000,
    // Tiempo que se guarda el catálogo de servicios en memoria
    CACHE_SERVICIOS: 5 * 60 * 1000
};

/**
 * Construye una URL completa de la API
 * @param {string} ruta - Ruta relativa (ej. '/api/turnos')
 * @returns {string}
 */
export const urlAPI = (ruta) => `${CONFIG.API_URL}${ruta}`;

/**
 * URL del flujo de eventos en tiempo real
 * @returns {string}
 */
export const urlEventos = () => urlAPI('/api/eventos');