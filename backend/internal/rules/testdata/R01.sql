-- fixture: R01 di atas db/seed/demo.sql, paket PS-07, @today = 2026-10-08
-- scenario: Menghapus baris CP 60-01 membuat detection control-nya tidak tercakup.
-- expect: [{"object_type": "controls", "field": "text", "params": {"charNo": "60-03", "control": "AOI 100%"}}, {"object_type": "controls", "field": "text", "params": {"charNo": "60-01", "control": "AOI 100% + ICT"}}]
DELETE FROM cp_lines WHERE id = (SELECT l.id FROM cp_lines l JOIN characteristics c ON c.id = l.characteristic_id WHERE l.package_id = (SELECT id FROM packages WHERE code = 'PS-07') AND c.char_no = '60-01' AND l.eval_technique = 'AOI + ICT');
