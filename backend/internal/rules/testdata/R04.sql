-- fixture: R04 di atas db/seed/demo.sql, paket PS-07, @today = 2026-10-08
-- scenario: D di step 90 diperbaiki; D terlalu optimis di step 50 (AOI hanya membenarkan D >= 3).
-- expect: [{"object_type": "failure_chains", "field": "d", "params": {"opNo": "50", "d": 2, "dMin": 3}}]
UPDATE failure_chains SET d = 6 WHERE id = (SELECT ch.id FROM failure_chains ch JOIN failure_causes fc ON fc.id = ch.failure_cause_id WHERE ch.package_id = (SELECT id FROM packages WHERE code = 'PS-07') AND fc.text = 'Inspection lighting insufficient');
UPDATE failure_chains SET d = 2 WHERE id = (SELECT ch.id FROM failure_chains ch JOIN failure_causes fc ON fc.id = ch.failure_cause_id WHERE ch.package_id = (SELECT id FROM packages WHERE code = 'PS-07') AND fc.text = 'Wrong reel loaded on feeder');
