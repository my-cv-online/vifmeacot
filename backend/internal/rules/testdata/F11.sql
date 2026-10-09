-- fixture: F11 di atas db/seed/demo.sql, paket PS-07, @today = 2026-10-08
-- scenario: Aksi kedua yang lewat target.
-- expect: [{"object_type": "actions", "field": "targetDate", "params": {"daysLate": 13}}, {"object_type": "actions", "field": "targetDate", "params": {"daysLate": 7}}]
UPDATE actions SET target_date = '2026-10-01' WHERE failure_chain_id = (SELECT ch.id FROM failure_chains ch JOIN failure_causes fc ON fc.id = ch.failure_cause_id WHERE ch.package_id = (SELECT id FROM packages WHERE code = 'PS-07') AND fc.text = 'Moisture absorbed by opened PCBs');
