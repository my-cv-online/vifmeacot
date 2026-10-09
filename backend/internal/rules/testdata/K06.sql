-- fixture: K06 di atas db/seed/demo.sql, paket PS-07, @today = 2026-10-08
-- scenario: Header PFMEA kini juga berbeda dengan data part (part number).
-- expect: [{"object_type": "documents", "field": "header", "params": {"docType": "CP", "docChangeLevel": "A"}}, {"object_type": "documents", "field": "header", "params": {"docType": "PFMEA", "docPartNo": "PS-08"}}]
UPDATE documents SET header = header || '{"partNo": "PS-08"}'::jsonb
WHERE package_id = (SELECT id FROM packages WHERE code = 'PS-07') AND doc_type = 'PFMEA';
