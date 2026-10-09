-- fixture: S01 di atas db/seed/demo.sql, paket PS-07, @today = 2026-10-08
-- scenario: Karakteristik SC baru yang belum ada di PFMEA maupun CP.
-- expect: [{"object_type": "characteristics", "field": "scSymbolId", "params": {"charNo": "60-05", "missingIn": "PFMEA"}}, {"object_type": "characteristics", "field": "scSymbolId", "params": {"charNo": "60-05", "missingIn": "Control Plan"}}]
INSERT INTO characteristics (package_id, step_id, char_no, kind, name, spec, sc_symbol_id)
VALUES ((SELECT id FROM packages WHERE code = 'PS-07'), (SELECT id FROM process_steps WHERE package_id = (SELECT id FROM packages WHERE code = 'PS-07') AND op_no = '60'), '60-05', 'product', 'Coplanarity', '<= 0.1 mm', (SELECT id FROM sc_symbols WHERE code = 'SC'));
