-- fixture: F12 di atas db/seed/demo.sql, paket PS-07, @today = 2026-10-08
-- scenario: Nomor dokumen diketik di sel kontrol.
-- expect: [{"object_type": "failure_causes", "field": "text", "params": {"term": "operator error"}}, {"object_type": "controls", "field": "text", "params": {"term": "WI-"}}]
UPDATE controls SET text = 'Visual inspection 100% per WI-QC-090'
WHERE failure_chain_id = (SELECT ch.id FROM failure_chains ch JOIN failure_causes fc ON fc.id = ch.failure_cause_id WHERE ch.package_id = (SELECT id FROM packages WHERE code = 'PS-07') AND fc.text = 'Inspection lighting insufficient') AND kind = 'detection';
