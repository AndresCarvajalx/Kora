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

const DIGITOS = ['cero', 'uno', 'dos', 'tres', 'cuatro', 'cinco', 'seis', 'siete', 'ocho', 'nueve'];

/**
 * Texto que se dice en voz alta, dígito por dígito y con los ceros:
 * "MED-004" → "Turno, M, E, D, cero, cero, cuatro"
 * @param {string} numero
 * @returns {string}
 */
export const textoAnuncio = (numero) => {
    const [prefijo, consecutivo = ''] = String(numero ?? '').split('-');
    // Las comas hacen una pausa ligera entre cada letra y cada dígito
    const letras = (prefijo || '').split('').join(', ');
    const cifras = consecutivo.split('').map((c) => DIGITOS[c] ?? c).join(', ');
    return `Turno, ${letras}, ${cifras}`.replace(/,\s*$/, '');
};

/**
 * Voz del anuncio: Microsoft Jorge Online (Natural). Si el navegador no la tiene
 * (p. ej. no es Edge), se usa la mejor voz en español disponible.
 * @returns {SpeechSynthesisVoice|undefined}
 */
const elegirVozAnuncio = () => {
    const voces = window.speechSynthesis.getVoices();
    const espanol = voces.filter((v) => v.lang.toLowerCase().startsWith('es'));

    return espanol.find((v) => /jorge/i.test(v.name) && /natural/i.test(v.name))
        || espanol.find((v) => /natural/i.test(v.name))
        || espanol.find((v) => !v.localService)
        || espanol[0];
};

/**
 * Dice un texto en voz alta
 * @param {string} texto
 * @param {number} [retardo] milisegundos de espera antes de hablar
 */
export const hablar = (texto, retardo = 0) => {
    if (!('speechSynthesis' in window) || !texto) return;

    const mensaje = new SpeechSynthesisUtterance(texto);
    const voz = elegirVozAnuncio();

    mensaje.lang = voz?.lang || 'es-MX';
    if (voz) mensaje.voice = voz;
    mensaje.rate = 0.78;

    window.speechSynthesis.cancel();
    window.setTimeout(() => window.speechSynthesis.speak(mensaje), retardo);
};

/**
 * Dice el número de turno en voz alta
 * @param {string} numero
 */
export const anunciarTurno = (numero) => {
    if (!activo || !numero) return;
    // Se espera a que termine el timbre para que no se pisen
    hablar(textoAnuncio(numero), DURACION_AVISO_MS - 200);
};

/**
 * Una nota de campana: fundamental + armónicos con caída suave
 */
const tocarCampana = (ctx, frecuencia, inicio, duracion, volumenBase) => {
    const armonicos = [
        { razon: 1, ganancia: 1 },
        { razon: 2, ganancia: 0.4 },
        { razon: 3, ganancia: 0.18 },
        { razon: 4.2, ganancia: 0.08 }
    ];

    armonicos.forEach(({ razon, ganancia }) => {
        const oscilador = ctx.createOscillator();
        const volumen = ctx.createGain();

        oscilador.type = 'sine';
        oscilador.frequency.value = frecuencia * razon;

        // Ataque rápido y caída exponencial; los agudos se apagan antes
        const pico = volumenBase * ganancia;
        const fin = inicio + duracion / Math.sqrt(razon);
        volumen.gain.setValueAtTime(0.0001, inicio);
        volumen.gain.exponentialRampToValueAtTime(pico, inicio + 0.015);
        volumen.gain.exponentialRampToValueAtTime(0.0001, fin);

        oscilador.connect(volumen);
        volumen.connect(ctx.destination);
        oscilador.start(inicio);
        oscilador.stop(fin + 0.05);
    });
};

/** Duración aproximada del aviso, en milisegundos */
export const DURACION_AVISO_MS = 1400;

/**
 * Reproduce un timbre de tres campanas ascendentes (Do-Mi-Sol) como aviso de llamado
 */
export const reproducirAviso = () => {
    if (!activo) return;

    const ctx = obtenerContexto();
    if (!ctx) return;

    const ahora = ctx.currentTime;
    const notas = [
        { frecuencia: 523.25, retardo: 0, duracion: 0.9 },    // Do
        { frecuencia: 659.25, retardo: 0.28, duracion: 0.9 }, // Mi
        { frecuencia: 783.99, retardo: 0.56, duracion: 1.4 }  // Sol
    ];

    notas.forEach(({ frecuencia, retardo, duracion }) => {
        tocarCampana(ctx, frecuencia, ahora + retardo, duracion, 0.22);
    });
};
