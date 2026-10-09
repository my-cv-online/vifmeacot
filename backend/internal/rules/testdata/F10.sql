-- fixture: F10 di atas db/seed/demo.sql, paket PS-07, @today = 2026-10-08
-- scenario: S baru lebih rendah dari S saat ini tanpa catatan perubahan desain.
-- expect: [{"object_type": "actions", "field": "designChangeNote", "params": {"s": 5, "newS": 4}}]
UPDATE actions SET new_s = 4 WHERE failure_chain_id = (SELECT ch.id FROM failure_chains ch JOIN failure_causes fc ON fc.id = ch.failure_cause_id WHERE ch.package_id = (SELECT id FROM packages WHERE code = 'PS-07') AND fc.text = 'Moisture absorbed by opened PCBs');
