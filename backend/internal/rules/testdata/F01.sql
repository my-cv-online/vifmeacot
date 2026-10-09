-- fixture: F01 di atas db/seed/demo.sql, paket PS-07, @today = 2026-10-08
-- scenario: Teks failure mode kosong dan D tidak diisi.
-- expect: [{"object_type": "failure_modes", "field": "text", "params": {"label": "Failure mode", "opNo": "50"}}, {"object_type": "failure_chains", "field": "d", "params": {"label": "Detection (D)", "opNo": "50"}}]
UPDATE failure_modes SET text = '' WHERE id = (SELECT id FROM failure_modes WHERE package_id = (SELECT id FROM packages WHERE code = 'PS-07') AND text = 'Component polarity reversed');
UPDATE failure_chains SET d = NULL WHERE id = (SELECT ch.id FROM failure_chains ch JOIN failure_causes fc ON fc.id = ch.failure_cause_id WHERE ch.package_id = (SELECT id FROM packages WHERE code = 'PS-07') AND fc.text = 'Wrong reel loaded on feeder');
