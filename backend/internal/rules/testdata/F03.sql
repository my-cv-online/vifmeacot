-- fixture: F03 di atas db/seed/demo.sql, paket PS-07, @today = 2026-10-08
-- scenario: Teks effect yang sama dinilai 6 sekali dan 7 dua kali.
-- expect: [{"object_type": "failure_effects", "field": "s", "params": {"s": 7, "values": "6, 7"}}, {"object_type": "failure_effects", "field": "s", "params": {"s": 7, "values": "6, 7"}}, {"object_type": "failure_effects", "field": "s", "params": {"s": 6, "values": "6, 7"}}]
UPDATE failure_effects SET text = 'Board fails ICT/FCT; rework or scrap', s = 6
WHERE failure_mode_id = (SELECT id FROM failure_modes WHERE package_id = (SELECT id FROM packages WHERE code = 'PS-07') AND text = 'Missing component');
