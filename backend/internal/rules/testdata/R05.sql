-- fixture: R05 di atas db/seed/demo.sql, paket PS-07, @today = 2026-10-08
-- scenario: O = 3 dengan satu-satunya prevention control dihapus.
-- expect: [{"object_type": "failure_chains", "field": "o", "params": {"o": 3, "opNo": "60"}}]
DELETE FROM controls WHERE failure_chain_id = (SELECT ch.id FROM failure_chains ch JOIN failure_causes fc ON fc.id = ch.failure_cause_id WHERE ch.package_id = (SELECT id FROM packages WHERE code = 'PS-07') AND fc.text = 'Reflow profile out of window') AND kind = 'prevention';
