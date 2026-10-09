-- fixture: S03 di atas db/seed/demo.sql, paket PS-07, @today = 2026-10-08
-- scenario: Menandai 90-01 sebagai SC memperlihatkan kontrolnya yang lemah (visual 100%).
-- expect: [{"object_type": "cp_lines", "field": "controlMethod", "params": {"charNo": "90-01", "symbol": "SC"}}]
UPDATE characteristics SET sc_symbol_id = (SELECT id FROM sc_symbols WHERE code = 'SC')
WHERE id = (SELECT id FROM characteristics WHERE package_id = (SELECT id FROM packages WHERE code = 'PS-07') AND char_no = '90-01');
