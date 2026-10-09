-- fixture: K01 di atas db/seed/demo.sql, paket PS-07, @today = 2026-10-08
-- scenario: Step baru tanpa failure mode dilaporkan.
-- expect: [{"object_type": "process_steps", "field": "name", "params": {"opNo": "95"}}]
INSERT INTO process_steps (package_id, op_no, seq, name, symbol)
VALUES ((SELECT id FROM packages WHERE code = 'PS-07'), '95', 95, 'Conformal coating', 'operation');
