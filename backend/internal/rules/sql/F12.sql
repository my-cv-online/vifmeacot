-- rule: F12
-- reads: failure_modes, failure_effects, failure_causes, controls, actions, banned_terms
-- message: Text contains the term "{{.term}}": {{.reason}}{{if .suggestion}} Suggestion: {{.suggestion}}{{end}}
-- returns: object_type, object_id, field, key, params (lihat docs/06-rules.md)
WITH t AS (
  SELECT 'failure_causes'::text AS object_type, fc.id AS object_id, 'text'::text AS field, fc.text AS txt, 'cause'::text AS kind
  FROM failure_causes fc WHERE fc.package_id = @package_id
  UNION ALL
  SELECT 'controls', k.id, 'text', k.text, 'control' FROM controls k WHERE k.package_id = @package_id
  UNION ALL
  SELECT 'failure_modes', fm.id, 'text', fm.text, 'failure_mode' FROM failure_modes fm WHERE fm.package_id = @package_id
  UNION ALL
  SELECT 'failure_effects', e.id, 'text', e.text, 'effect' FROM failure_effects e WHERE e.package_id = @package_id
  UNION ALL
  SELECT 'actions', a.id, 'text', a.text, 'action' FROM actions a WHERE a.package_id = @package_id
)
SELECT t.object_type, t.object_id, t.field, b.id::text,
       jsonb_build_object('term', b.term, 'reason', b.reason, 'suggestion', coalesce(b.suggestion, ''))
FROM t
JOIN banned_terms b ON b.is_active
                   AND (b.applies_to = 'any' OR b.applies_to = t.kind)
                   AND strpos(lower(t.txt), lower(b.term)) > 0
