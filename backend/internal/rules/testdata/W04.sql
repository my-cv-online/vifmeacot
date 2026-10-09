-- fixture: W04 di atas db/seed/demo.sql, paket PS-07, @today = 2026-10-08
-- scenario: Review terbaru menghilangkan temuan demo.
-- expect: []
UPDATE packages SET last_reviewed_at = '2026-06-01 09:00+07' WHERE id = (SELECT id FROM packages WHERE code = 'PS-07');
