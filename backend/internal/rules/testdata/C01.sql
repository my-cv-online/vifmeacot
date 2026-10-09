-- fixture: C01 di atas db/seed/demo.sql, paket PS-07, @today = 2026-10-08
-- scenario: Sample size dan reaction plan kosong pada baris CP step 90.
-- expect: [{"object_type": "cp_lines", "field": "sampleSize", "params": {"charNo": "90-01"}}, {"object_type": "cp_lines", "field": "reactionPlan", "params": {"charNo": "90-01"}}]
UPDATE cp_lines SET sample_size = '' WHERE id = (SELECT l.id FROM cp_lines l JOIN characteristics c ON c.id = l.characteristic_id WHERE l.package_id = (SELECT id FROM packages WHERE code = 'PS-07') AND c.char_no = '90-01' AND l.eval_technique = 'Visual');
UPDATE reaction_plans SET text = '' WHERE cp_line_id = (SELECT l.id FROM cp_lines l JOIN characteristics c ON c.id = l.characteristic_id WHERE l.package_id = (SELECT id FROM packages WHERE code = 'PS-07') AND c.char_no = '90-01' AND l.eval_technique = 'Visual');
