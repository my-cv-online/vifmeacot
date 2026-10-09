-- fixture: S02 di atas db/seed/demo.sql, paket PS-07, @today = 2026-10-08
-- scenario: Karakteristik CC kedua di paket Customer B (Customer B tidak punya pemetaan CC).
-- expect: [{"object_type": "characteristics", "field": "scSymbolId", "params": {"charNo": "60-02", "symbol": "CC"}}, {"object_type": "characteristics", "field": "scSymbolId", "params": {"charNo": "50-01", "symbol": "CC"}}]
UPDATE characteristics SET sc_symbol_id = (SELECT id FROM sc_symbols WHERE code = 'CC')
WHERE id = (SELECT id FROM characteristics WHERE package_id = (SELECT id FROM packages WHERE code = 'PS-07') AND char_no = '50-01');
