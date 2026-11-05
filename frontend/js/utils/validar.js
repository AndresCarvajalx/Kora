// Validaciones y normalización de los datos del paciente

const RE_NOMBRE = /^\p{L}[\p{L}\s'.-]*$/u;
const RE_DOCUMENTO = /^\d{5,15}$/;

/** Quita espacios repetidos de un texto */
export const compactar = (texto) => String(texto ?? '').replace(/\s+/g, ' ').trim();

/**
 * Deja un teléfono solo con dígitos (conserva un "+" inicial).
 * "300 123-4567" → "3001234567"
 * @param {string} valor
 * @returns {string}
 */
export const normalizarContacto = (valor) => {
    const limpio = String(valor ?? '').trim().replace(/[\s().-]/g, '');
    return limpio;
};

/**
 * @param {string} valor
 * @returns {boolean} Si el texto parece un teléfono (y no un número de turno)
 */
export const pareceTelefono = (valor) => /^\+?[\d\s().-]+$/.test(String(valor ?? '').trim());

/**
 * Cada validador devuelve un mensaje de error o '' si el valor es correcto.
 */
export const validarNombre = (valor) => {
    const nombre = compactar(valor);
    if (!nombre) return 'Ingresa tu nombre completo';
    if (nombre.length < 3) return 'El nombre es demasiado corto';
    if (nombre.length > 80) return 'El nombre no puede superar 80 caracteres';
    if (!RE_NOMBRE.test(nombre)) return 'Usa solo letras, sin números ni símbolos';
    if (nombre.split(' ').filter((parte) => parte.length >= 2).length < 2) {
        return 'Escribe al menos nombre y apellido';
    }
    return '';
};

export const validarDocumento = (valor) => {
    const documento = String(valor ?? '').replace(/[\s.-]/g, '');
    if (!documento) return ''; // opcional
    if (!/^\d+$/.test(documento)) return 'El documento solo puede tener números';
    if (!RE_DOCUMENTO.test(documento)) return 'Debe tener entre 5 y 15 dígitos';
    return '';
};

export const validarContacto = (valor) => {
    const contacto = normalizarContacto(valor);
    if (!contacto) return 'Ingresa un número de contacto';
    if (!/^\+?\d+$/.test(contacto)) return 'Solo se permiten números';
    const digitos = contacto.replace('+', '');
    if (digitos.length < 7 || digitos.length > 15) {
        return 'El número debe tener entre 7 y 15 dígitos';
    }
    return '';
};

export const validarServicio = (valor) => (valor ? '' : 'Selecciona un servicio médico');

export const validarTipoAtencion = (valor, requerido) =>
    (requerido && !valor ? 'Selecciona el tipo de atención' : '');
