-- fixture: F09 di atas db/seed/demo.sql, paket PS-07, @today = 2026-10-08
-- scenario: Aksi selesai tanpa D baru.
-- expect: [{"object_type": "actions", "field": "newD", "params": {"label": "new D"}}]
UPDATE actions SET status = 'done', completed_on = '2026-10-01', action_taken = 'Nozzle replaced', new_s = 7, new_o = 2
WHERE failure_chain_id = (SELECT ch.id FROM failure_chains ch JOIN failure_causes fc ON fc.id = ch.failure_cause_id WHERE ch.package_id = (SELECT id FROM packages WHERE code = 'PS-07') AND fc.text = 'Nozzle clogged; pick-up error');
