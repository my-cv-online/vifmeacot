-- fixture: T03 di atas db/seed/demo.sql, paket PS-07, @today = 2026-10-08
-- scenario: Dua konflik sinkron yang belum diselesaikan: konflik override kolom pada baris CP 20-02 dan konflik baris pada step 20 (dihapus di template, masih dipakai baris lokal).
-- expect: [{"object_type": "cp_lines", "field": "sampleFreq", "params": {"column": "sampleFreq", "rev": 2, "localValue": "Every 3 reels", "templateValue": "Every reel"}}, {"object_type": "process_steps", "field": "", "params": {"column": "", "rev": 2, "label": "20 IQC"}}]
INSERT INTO sync_runs (id, general_package_id, from_rev, to_rev, status, packages_total, created_by)
VALUES ('00000000-0000-7000-8000-00000000f003', (SELECT id FROM packages WHERE code = 'GENERAL'), 1, 2, 'done', 1, (SELECT id FROM users WHERE username = 'rsaputri'));
INSERT INTO sync_changes (sync_run_id, package_id, table_name, row_id, column_name, old_value, new_value, action)
VALUES ('00000000-0000-7000-8000-00000000f003', (SELECT id FROM packages WHERE code = 'PS-07'), 'cp_lines',
        (SELECT l.id FROM cp_lines l JOIN characteristics c ON c.id = l.characteristic_id
          WHERE l.package_id = (SELECT id FROM packages WHERE code = 'PS-07') AND c.char_no = '20-02'),
        'sample_freq', '"Every 3 reels"', '"Every reel"', 'conflict');
INSERT INTO sync_changes (sync_run_id, package_id, table_name, row_id, column_name, old_value, new_value, action)
VALUES ('00000000-0000-7000-8000-00000000f003', (SELECT id FROM packages WHERE code = 'PS-07'), 'process_steps',
        (SELECT s.id FROM process_steps s WHERE s.package_id = (SELECT id FROM packages WHERE code = 'PS-07') AND s.op_no = '20'),
        NULL, '{"label": "20 IQC"}', NULL, 'conflict');
