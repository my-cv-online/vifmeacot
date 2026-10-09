-- fixture: R06 di atas db/seed/demo.sql, paket PS-07, @today = 2026-10-08
-- scenario: Menambah frekuensi verifikasi menghilangkan temuan demo.
-- expect: []
UPDATE cp_lines SET ep_verify_freq = 'Start of shift, with a wrong reel' WHERE id = (SELECT l.id FROM cp_lines l JOIN characteristics c ON c.id = l.characteristic_id WHERE l.package_id = (SELECT id FROM packages WHERE code = 'PS-07') AND c.char_no = '50-01' AND l.eval_technique = 'Feeder interlock');
