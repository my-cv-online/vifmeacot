-- fixture: R03 di atas db/seed/demo.sql, paket PS-07, @today = 2026-10-08
-- scenario: Memutus tautan baris CP AOI dari PFMEA menambah R03 kedua.
-- expect: [{"object_type": "cp_lines", "field": "controlId", "params": {"charNo": "75-01"}}, {"object_type": "cp_lines", "field": "controlId", "params": {"charNo": "70-01"}}]
UPDATE cp_lines SET control_id = NULL, failure_mode_id = NULL WHERE id = (SELECT l.id FROM cp_lines l JOIN characteristics c ON c.id = l.characteristic_id WHERE l.package_id = (SELECT id FROM packages WHERE code = 'PS-07') AND c.char_no = '70-01' AND l.eval_technique = 'Golden sample');
