-- ============================================================
-- KORA · Migración 003 · Catálogo inicial de servicios
-- ============================================================

INSERT INTO servicios (codigo, nombre, descripcion, duracion_minutos, orden) VALUES
    ('MED', 'Medicina General', 'Atención médica general para adultos y niños.', 15, 1),
    ('PED', 'Pediatría', 'Atención especializada en niños y adolescentes.', 15, 2),
    ('CAR', 'Cardiología', 'Evaluación cardiovascular y estudios del corazón.', 20, 3),
    ('ODO', 'Odontología', 'Odontología general y tratamientos dentales.', 30, 4),
    ('OFT', 'Oftalmología', 'Valoración de la visión y salud ocular.', 20, 5),
    ('PSI', 'Psicología', 'Atención psicológica y terapia individual.', 40, 6)
ON CONFLICT (codigo) DO NOTHING;

INSERT INTO tipos_atencion (servicio_id, nombre, orden)
SELECT s.id, t.nombre, t.orden
FROM servicios s
JOIN (VALUES
    ('MED', 'Consulta general',   1),
    ('MED', 'Control',            2),
    ('MED', 'Vacunación',         3),
    ('PED', 'Consulta pediátrica', 1),
    ('PED', 'Control de crecimiento', 2),
    ('PED', 'Vacunación',         3),
    ('CAR', 'Consulta',           1),
    ('CAR', 'Control',            2),
    ('CAR', 'Ecocardiograma',     3),
    ('CAR', 'Prueba de esfuerzo', 4),
    ('ODO', 'Consulta',           1),
    ('ODO', 'Limpieza dental',    2),
    ('ODO', 'Endodoncia',          3),
    ('ODO', 'Ortodoncia',          4),
    ('OFT', 'Consulta',           1),
    ('OFT', 'Valoración visual',  2),
    ('OFT', 'Lentes',             3),
    ('PSI', 'Consulta',           1),
    ('PSI', 'Terapia',            2),
    ('PSI', 'Orientación',        3)
) AS t(codigo, nombre, orden) ON t.codigo = s.codigo
WHERE NOT EXISTS (
    SELECT 1 FROM tipos_atencion ta
    WHERE ta.servicio_id = s.id AND ta.nombre = t.nombre
);