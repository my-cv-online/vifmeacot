-- fixture: T04 di atas db/seed/demo.sql, paket PS-07, @today = 2026-10-08
-- scenario: Baris tertaut dilepas dari template tanpa alasan.
-- expect: [{"object_type": "cp_lines", "field": "detachReason", "params": {}}]
UPDATE cp_lines SET sync_status = 'detached', detach_reason = NULL
WHERE package_id = (SELECT id FROM packages WHERE code = 'PS-07') AND id = (SELECT l.id FROM cp_lines l JOIN characteristics c ON c.id = l.characteristic_id
                                   WHERE l.package_id = (SELECT id FROM packages WHERE code = 'PS-07') AND c.char_no = '20-02');
