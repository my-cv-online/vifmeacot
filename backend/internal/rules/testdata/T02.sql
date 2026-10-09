-- fixture: T02 di atas db/seed/demo.sql, paket PS-07, @today = 2026-10-08
-- scenario: Template General merilis rev 2 tetapi paket model masih mengikuti rev 1.
-- expect: [{"object_type": "packages", "field": "templateRev", "params": {"syncedRev": 1, "latestRev": 2}}]
UPDATE packages SET revision = 2 WHERE id = (SELECT id FROM packages WHERE code = 'GENERAL');
UPDATE package_links SET state = 'pending' WHERE model_package_id = (SELECT id FROM packages WHERE code = 'PS-07');
