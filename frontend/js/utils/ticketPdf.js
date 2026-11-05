// Genera el PDF del turno con formato de ticket de impresora térmica (80 mm)
import { formatearEspera, formatearFechaHora, nombrePaciente } from './format.js';

const ANCHO = 80;
const ALTO = 150;
const MARGEN = 6;

/**
 * Crea el PDF del ticket y lo descarga.
 * @param {Object} ticket
 * @returns {boolean} false si la librería de PDF no está disponible
 */
export const descargarTicketPdf = (ticket) => {
    const JsPDF = window.jspdf?.jsPDF;
    if (!JsPDF || !ticket) return false;

    const doc = new JsPDF({ unit: 'mm', format: [ANCHO, ALTO] });
    const centro = ANCHO / 2;
    let y = 10;

    const linea = () => {
        doc.setDrawColor(150);
        doc.setLineDashPattern([1, 1], 0);
        doc.line(MARGEN, y, ANCHO - MARGEN, y);
        doc.setLineDashPattern([], 0);
        y += 5;
    };

    const campo = (etiqueta, valor) => {
        doc.setFont('helvetica', 'normal');
        doc.setFontSize(7);
        doc.setTextColor(110);
        doc.text(etiqueta.toUpperCase(), MARGEN, y);
        y += 3.6;
        doc.setFont('helvetica', 'bold');
        doc.setFontSize(9.5);
        doc.setTextColor(0);
        const partes = doc.splitTextToSize(String(valor || '—'), ANCHO - MARGEN * 2);
        doc.text(partes, MARGEN, y);
        y += partes.length * 4.2 + 1.8;
    };

    // Encabezado
    doc.setFont('helvetica', 'bold');
    doc.setFontSize(15);
    doc.setTextColor(0);
    doc.text('KORA', centro, y, { align: 'center' });
    y += 5;
    doc.setFont('helvetica', 'normal');
    doc.setFontSize(8);
    doc.text('Sistema de gestión de turnos', centro, y, { align: 'center' });
    y += 5;
    linea();

    // Número de turno
    doc.setFontSize(8);
    doc.text('TU TURNO', centro, y, { align: 'center' });
    y += 11;
    doc.setFont('helvetica', 'bold');
    doc.setFontSize(34);
    doc.text(ticket.numero_turno || '—', centro, y, { align: 'center' });
    y += 6;
    linea();

    campo('Servicio', ticket.servicio_nombre);
    campo('Tipo de atención', ticket.tipo_atencion_nombre || 'General');
    campo('Paciente', nombrePaciente(ticket));
    if (ticket.cliente_documento) campo('Documento', ticket.cliente_documento);

    if (ticket.estado === 'PENDIENTE') {
        const antes = ticket.personas_antes ?? 0;
        campo('Personas antes de ti', String(antes));
        campo('Espera estimada', formatearEspera(ticket.tiempo_estimado_min));
    }

    campo('Emitido', formatearFechaHora(ticket.creado_en));
    campo('Vence', formatearFechaHora(ticket.expira_en));
    linea();

    doc.setFont('helvetica', 'normal');
    doc.setFontSize(7.5);
    doc.setTextColor(60);
    doc.text('Espera a que llamen tu número en el tablero.', centro, y, { align: 'center' });
    y += 3.6;
    doc.text('Conserva este ticket para consultar tu turno.', centro, y, { align: 'center' });

    doc.save(`turno-${ticket.numero_turno || 'kora'}.pdf`);
    return true;
};
