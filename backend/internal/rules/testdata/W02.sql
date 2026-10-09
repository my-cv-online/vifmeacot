-- fixture: W02 di atas db/seed/demo.sql, paket PS-07, @today = 2026-10-08
-- scenario: Final inspection tanpa cabang NG.
-- expect: [{"object_type": "process_steps", "field": "ngFlow", "params": {"opNo": "90"}}]
DELETE FROM step_flows WHERE package_id = (SELECT id FROM packages WHERE code = 'PS-07') AND from_step_id = (SELECT id FROM process_steps WHERE package_id = (SELECT id FROM packages WHERE code = 'PS-07') AND op_no = '90') AND kind = 'ng';
