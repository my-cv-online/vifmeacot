-- fixture: K05 di atas db/seed/demo.sql, paket PS-07, @today = 2026-10-08
-- scenario: Mesin tak terdaftar kedua pada baris CP.
-- expect: [{"object_type": "cp_lines", "field": "machines", "params": {"opNo": "90", "unknown": "Lux meter"}}, {"object_type": "cp_lines", "field": "machines", "params": {"opNo": "50", "unknown": "Glue dispenser"}}]
UPDATE cp_lines SET machines = ARRAY['Mounter', 'Glue dispenser'] WHERE id = (SELECT l.id FROM cp_lines l JOIN characteristics c ON c.id = l.characteristic_id WHERE l.package_id = (SELECT id FROM packages WHERE code = 'PS-07') AND c.char_no = '50-01' AND l.eval_technique = 'AOI');
