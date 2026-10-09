-- fixture: T01 di atas db/seed/demo.sql, paket PS-07, @today = 2026-10-08
-- scenario: Step Shipping tertaut (wajib di rev 1) dihapus dari paket model.
-- expect: [{"object_type": "packages", "field": "steps", "params": {"opNo": "110", "rev": 1}}]
DELETE FROM cp_lines WHERE step_id = (SELECT id FROM process_steps WHERE package_id = (SELECT id FROM packages WHERE code = 'PS-07') AND op_no = '110');
DELETE FROM failure_modes WHERE step_id = (SELECT id FROM process_steps WHERE package_id = (SELECT id FROM packages WHERE code = 'PS-07') AND op_no = '110');
DELETE FROM process_steps WHERE id = (SELECT id FROM process_steps WHERE package_id = (SELECT id FROM packages WHERE code = 'PS-07') AND op_no = '110');
