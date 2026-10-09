-- fixture: F07 di atas db/seed/demo.sql, paket PS-07, @today = 2026-10-08
-- scenario: S dinaikkan menjadi 9 pada chain tanpa aksi atau justifikasi.
-- expect: [{"object_type": "failure_chains", "field": "s", "params": {"s": 9, "opNo": "60"}}]
UPDATE failure_effects SET s = 9 WHERE failure_mode_id = (SELECT id FROM failure_modes WHERE package_id = (SELECT id FROM packages WHERE code = 'PS-07') AND text = 'Component damaged by overheating');
