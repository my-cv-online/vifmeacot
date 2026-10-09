-- fixture: W01 di atas db/seed/demo.sql, paket PS-07, @today = 2026-10-08
-- scenario: Menganalisis step rework menghilangkan temuan demo.
-- expect: []
INSERT INTO failure_modes (package_id, step_id, characteristic_id, text)
VALUES ((SELECT id FROM packages WHERE code = 'PS-07'), (SELECT id FROM process_steps WHERE package_id = (SELECT id FROM packages WHERE code = 'PS-07') AND op_no = '75'), (SELECT id FROM characteristics WHERE package_id = (SELECT id FROM packages WHERE code = 'PS-07') AND char_no = '75-01'), 'Joint damaged by excessive heat during rework');
