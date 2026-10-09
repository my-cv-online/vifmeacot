-- fixture: K03 di atas db/seed/demo.sql, paket PS-07, @today = 2026-10-08
-- scenario: Failure mode step 50 yang menunjuk karakteristik milik step 60.
-- expect: [{"object_type": "failure_modes", "field": "characteristicId", "params": {"opNo": "50", "charNo": "60-01", "charOpNo": "60"}}]
UPDATE failure_modes SET characteristic_id = (SELECT id FROM characteristics WHERE package_id = (SELECT id FROM packages WHERE code = 'PS-07') AND char_no = '60-01') WHERE id = (SELECT id FROM failure_modes WHERE package_id = (SELECT id FROM packages WHERE code = 'PS-07') AND text = 'Wrong component placed');
