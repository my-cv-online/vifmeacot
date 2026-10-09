-- rule: F10
-- reads: actions, failure_chains, failure_effects
-- message: S lowered from {{.s}} to {{.newS}} after action "{{.action}}" without a design change note.
-- returns: object_type, object_id, field, key, params (lihat docs/06-rules.md)
SELECT 'actions'::text, a.id, 'designChangeNote'::text, ''::text,
       jsonb_build_object('s', ch.s, 'newS', a.new_s, 'action', a.text)
FROM actions a
JOIN failure_chains ch ON ch.id = a.failure_chain_id
WHERE a.package_id = @package_id
  AND a.new_s IS NOT NULL AND ch.s IS NOT NULL
  AND a.new_s < ch.s
  AND coalesce(btrim(a.design_change_note), '') = ''
