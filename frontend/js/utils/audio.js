// Reproductor de aviso sonoro al llamado de un turno (Web Audio API)
let contexto = null;
let activo = false;

const obtenerContexto = () => {
    if (!contexto) {
        const Constructor = window.AudioContext || window.webkitAudioContext;
        if (!Constructor) return null;
        contexto = new Constructor();
    }
    if (contexto.state === 'suspended') contexto.resume();
    return contexto;
};

/**
 * Activa o desactiva el aviso sonoro
 * @param {boolean} valor
 * @returns {boolean} Estado resultante
 */
export const activarSonido = (valor) => {
    activo = Boolean(valor);
    if (activo) obtenerContexto();
    return activo;
};

/**
 * @returns {boolean} Si el aviso sonoro está activo
 */
export const sonidoActivo = () => activo;

/**
 * Reproduce dos tonos cortos como aviso de llamado
 */
export const reproducirAviso = () => {
    if (!activo) return;

    const ctx = obtenerContexto();
    if (!ctx) return;

    const ahora = ctx.currentTime;

    [0, 0.22].forEach((retardo, indice) => {
        const oscilador = ctx.createOscillator();
        const volumen = ctx.createGain();

        oscilador.type = 'sine';
        oscilador.frequency.value = indice === 0 ? 880 : 660;

        volumen.gain.setValueAtTime(0.0001, ahora + retardo);
        volumen.gain.exponentialRampToValueAtTime(0.25, ahora + retardo + 0.02);
        volumen.gain.exponentialRampToValueAtTime(0.0001, ahora + retardo + 0.18);

        oscilador.connect(volumen);
        volumen.connect(ctx.destination);
        oscilador.start(ahora + retardo);
        oscilador.stop(ahora + retardo + 0.2);
    });
};