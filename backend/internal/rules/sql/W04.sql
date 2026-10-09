-- rule: W04
-- reads: packages, app_settings
-- message: Package last reviewed on {{.lastReviewedAt}}, more than {{.months}} months ago.
-- returns: object_type, object_id, field, key, params (lihat docs/06-rules.md)
WITH cfg AS (
  SELECT coalesce((SELECT (value)::int FROM app_settings WHERE key = 'review_interval_months'), 12) AS months
)
SELECT 'packages'::text, pk.id, 'lastReviewedAt'::text, ''::text,
       jsonb_build_object('lastReviewedAt', to_char(pk.last_reviewed_at, 'DD Mon YYYY'), 'months', cfg.months)
FROM packages pk, cfg
WHERE pk.id = @package_id
  AND pk.last_reviewed_at < (@today::date - make_interval(months => cfg.months))
