-- fixture: S04 di atas db/seed/demo.sql, paket PS-07, @today = 2026-10-08
-- scenario: Menurunkan syarat CC customer menjadi 8 menghilangkan temuan demo.
-- expect: []
UPDATE customers SET cc_min_severity = 8 WHERE code = 'CB';
