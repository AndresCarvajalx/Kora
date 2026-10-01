// Utilidades de manipulación del DOM

/**
 * Selecciona un único elemento
 * @param {string} selector
 * @param {ParentNode} [raiz]
 * @returns {Element|null}
 */
export const $ = (selector, raiz = document) => raiz.querySelector(selector);

/**
 * Selecciona varios elementos como arreglo
 * @param {string} selector
 * @param {ParentNode} [raiz]
 * @returns {Element[]}
 */
export const $$ = (selector, raiz = document) => Array.from(raiz.querySelectorAll(selector));

const ENTIDADES = {
    '&': '&amp;',
    '<': '&lt;',
    '>': '&gt;',
    '"': '&quot;',
    "'": '&#39;'
};

/**
 * Escapa texto antes de insertarlo como HTML
 * @param {unknown} valor
 * @returns {string}
 */
export const escapar = (valor) => String(valor ?? '').replace(/[&<>"']/g, (caracter) => ENTIDADES[caracter]);

/**
 * Inserta HTML en un contenedor por id
 * @param {string} idContenedor
 * @param {string} html
 * @returns {Element|null}
 */
export const renderHTML = (idContenedor, html) => {
    const contenedor = document.getElementById(idContenedor);
    if (!contenedor) return null;
    contenedor.innerHTML = html;
    return contenedor;
};

/**
 * Registra un listener mediante delegación de eventos
 * @param {ParentNode} raiz
 * @param {string} evento
 * @param {string} selector
 * @param {(detalle: { objetivo: Element, evento: Event }) => void} manejador
 */
export const delegar = (raiz, evento, selector, manejador) => {
    if (!raiz) return;
    raiz.addEventListener(evento, (eventoReal) => {
        const objetivo = eventoReal.target.closest(selector);
        if (objetivo && raiz.contains(objetivo)) {
            manejador({ objetivo, evento: eventoReal });
        }
    });
};

/**
 * Muestra una alerta temporal en un contenedor
 * @param {string} idContenedor
 * @param {'success'|'error'|'info'|'warning'} tipo
 * @param {string} mensaje
 */
export const mostrarAlerta = (idContenedor, tipo, mensaje) => {
    const contenedor = document.getElementById(idContenedor);
    if (!contenedor) return;

    contenedor.innerHTML = `<div class="alert alert-${tipo}">${escapar(mensaje)}</div>`;

    window.setTimeout(() => {
        const alerta = contenedor.querySelector('.alert');
        if (alerta) alerta.remove();
    }, 6000);
};

/**
 * Limpia el contenido de un contenedor de alertas
 * @param {string} idContenedor
 */
export const limpiarAlertas = (idContenedor) => {
    const contenedor = document.getElementById(idContenedor);
    if (contenedor) contenedor.innerHTML = '';
};

/**
 * Activa o desactiva el estado de carga de un botón
 * @param {string} idBoton
 * @param {boolean} cargando
 * @param {string} [textoCargando]
 */
export const estadoCarga = (idBoton, cargando, textoCargando = 'Procesando…') => {
    const boton = document.getElementById(idBoton);
    if (!boton) return;

    if (cargando) {
        if (!boton.dataset.textoOriginal) {
            boton.dataset.textoOriginal = boton.textContent.trim();
        }
        boton.disabled = true;
        boton.textContent = textoCargando;
        return;
    }

    boton.disabled = false;
    boton.textContent = boton.dataset.textoOriginal || 'Enviar';
    delete boton.dataset.textoOriginal;
};

/**
 * Marca o desmarca un campo con error de validación
 * @param {string} idCampo
 * @param {string} [mensaje]
 */
export const marcarErrorCampo = (idCampo, mensaje = '') => {
    const campo = document.getElementById(idCampo);
    if (!campo) return;

    campo.classList.toggle('has-error', Boolean(mensaje));

    const existente = campo.parentElement.querySelector('.field-error');
    if (existente) existente.remove();

    if (mensaje) {
        const aviso = document.createElement('span');
        aviso.className = 'field-error';
        aviso.textContent = mensaje;
        campo.parentElement.appendChild(aviso);
    }
};