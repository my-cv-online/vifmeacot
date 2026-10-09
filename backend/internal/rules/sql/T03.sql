-- rule: T03
-- reads: sync_changes
-- message: {{if .column}}Local override of {{.column}} conflicts with Template General rev {{.rev}}: local {{.localValue}}, template {{.templateValue}}.{{else}}Row {{.label}} was deleted from Template General rev {{.rev}} but local rows still use it. Choose "Detach from template" or "Follow template" (delete).{{end}}
-- returns: object_type, object_id, field, key, params (lihat docs/06-rules.md)
-- Konflik kolom: column_name terisi, old_value = nilai lokal, new_value = nilai template baru.
-- Konflik baris (template menghapus baris yang masih dipakai baris lokal): column_name IS NULL,
-- old_value = {"label": "..."}, new_value IS NULL. Lihat docs/07-template-general.md.
SELECT sc.table_name, sc.row_id, coalesce(snake_to_camel(sc.column_name), ''), sc.id::text,
       jsonb_build_object('column', coalesce(snake_to_camel(sc.column_name), ''), 'rev', r.to_rev,
                          'label', coalesce(sc.old_value ->> 'label', ''),
                          'localValue', coalesce(sc.old_value #>> '{}', '-'),
                          'templateValue', coalesce(sc.new_value #>> '{}', '-'))
FROM sync_changes sc
JOIN sync_runs r ON r.id = sc.sync_run_id
WHERE sc.package_id = @package_id
  AND sc.action = 'conflict'
  AND sc.resolved_at IS NULL
