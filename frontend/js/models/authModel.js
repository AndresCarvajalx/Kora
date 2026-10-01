// Modelo de Autenticación - Maneja login y token del personal
import { urlAPI } from '../config.js';

const TOKEN_KEY = 'kora_admin_token';

/**
 * Accede a localStorage sin fallar si el navegador lo bloquea
 * @param {'get'|'set'|'remove'} accion
 * @param {string} valor
 * @returns {string|null}
 */
const almacenamiento = (accion, valor) => {
    try {
        if (accion === 'set') return void window.localStorage.setItem(TOKEN_KEY, valor);
        if (accion === 'remove') return void window.localStorage.removeItem(TOKEN_KEY);
        return window.localStorage.getItem(TOKEN_KEY);
    } catch {
        return null;
    }
};

export const authModel = {
    /**
     * Inicia sesión del personal administrativo
     * @param {string} email - Email del administrador
     * @param {string} password - Contraseña
     * @returns {Promise<string>} Token de autenticación
     */
    async login(email, password) {
        const response = await fetch(urlAPI('/api/admin/login'), {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ email, password })
        });

        const data = await response.json().catch(() => ({}));

        if (!response.ok) {
            throw new Error(data.error || 'Error al iniciar sesión');
        }

        almacenamiento('set', data.data.token);

        return data.data.token;
    },

    /**
     * Obtiene el token almacenado
     * @returns {string|null}
     */
    getToken() {
        return almacenamiento('get');
    },

    /**
     * Verifica si hay una sesión activa
     * @returns {boolean}
     */
    estaAutenticado() {
        return this.getToken() !== null;
    },

    /**
     * Cierra sesión eliminando el token
     */
    logout() {
        almacenamiento('remove');
    },

    /**
     * Obtiene los headers de autenticación
     * @returns {Object} Headers con Authorization si hay sesión
     */
    getAuthHeaders() {
        const token = this.getToken();
        const headers = { 'Content-Type': 'application/json' };

        if (token) {
            headers.Authorization = `Bearer ${token}`;
        }

        return headers;
    }
};
