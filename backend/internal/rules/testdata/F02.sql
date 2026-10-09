-- fixture: F02 di atas db/seed/demo.sql, paket PS-07, @today = 2026-10-08
-- scenario: Dua failure mode diketik dalam satu sel.
-- expect: [{"object_type": "failure_modes", "field": "text", "params": {"text": "Burr / crack"}}]
UPDATE failure_modes SET text = 'Burr / crack' WHERE id = (SELECT id FROM failure_modes WHERE package_id = (SELECT id FROM packages WHERE code = 'PS-07') AND text = 'Cosmetic defect not detected');
