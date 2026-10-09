-- rule: T02
-- reads: package_links, packages
-- message: Package uses Template General rev {{.syncedRev}}; the latest is rev {{.latestRev}} (sync state: {{.state}}).
-- returns: object_type, object_id, field, key, params (lihat docs/06-rules.md)
SELECT 'packages'::text, pk.id, 'templateRev'::text, ''::text,
       jsonb_build_object('syncedRev', lk.synced_rev, 'latestRev', g.revision, 'state', lk.state)
FROM packages pk
JOIN package_links lk ON lk.model_package_id = pk.id
JOIN packages g ON g.id = lk.general_package_id
WHERE pk.id = @package_id
  AND lk.synced_rev < g.revision
  AND lk.state <> 'syncing'
