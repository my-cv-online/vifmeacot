-- fixture: R02 di atas db/seed/demo.sql, paket PS-07, @today = 2026-10-08
-- scenario: Menandai zone temperature alarm sebagai kontrol sistem menghilangkan temuan demo.
-- expect: []
UPDATE controls SET is_system_control = true, system_control_reason = 'Oven alarm connected to the andon'
WHERE package_id = (SELECT id FROM packages WHERE code = 'PS-07') AND text = 'Zone temperature alarm';
