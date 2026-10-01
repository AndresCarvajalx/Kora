// Entry point · Tablero de llamados en tiempo real
import { tableroController } from './controllers/tableroController.js';

document.addEventListener('DOMContentLoaded', () => {
    tableroController.init();
});