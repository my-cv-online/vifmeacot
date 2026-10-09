-- fixture: K02 di atas db/seed/demo.sql, paket PS-07, @today = 2026-10-08
-- scenario: Menghapus baris CP 50-03 menambah K02 kedua di samping baseline demo 60-03.
-- expect: [{"object_type": "characteristics", "field": "name", "params": {"charNo": "60-03"}}, {"object_type": "characteristics", "field": "name", "params": {"charNo": "50-03"}}]
DELETE FROM cp_lines WHERE id = (SELECT l.id FROM cp_lines l JOIN characteristics c ON c.id = l.characteristic_id WHERE l.package_id = (SELECT id FROM packages WHERE code = 'PS-07') AND c.char_no = '50-03' AND l.eval_technique = 'AOI');
