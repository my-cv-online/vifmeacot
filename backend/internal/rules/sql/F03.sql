-- rule: F03
-- reads: failure_effects
-- message: Effect "{{.effect}}" is rated with different S values in this package: {{.values}}.
-- returns: object_type, object_id, field, key, params (lihat docs/06-rules.md)
WITH e AS (
  SELECT e.id, e.s, e.text, lower(regexp_replace(btrim(e.text), '\s+', ' ', 'g')) AS norm
  FROM failure_effects e
  WHERE e.package_id = @package_id AND e.s IS NOT NULL AND btrim(e.text) <> ''
), g AS (
  SELECT norm, string_agg(DISTINCT s::text, ', ' ORDER BY s::text) AS svals
  FROM e GROUP BY norm HAVING count(DISTINCT s) > 1
)
SELECT 'failure_effects'::text, e.id, 's'::text, ''::text,
       jsonb_build_object('effect', e.text, 's', e.s, 'values', g.svals)
FROM e JOIN g USING (norm)
