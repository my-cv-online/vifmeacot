-- rule: F11
-- reads: actions
-- message: Action "{{.action}}" is {{.daysLate}} days overdue (target {{.targetDate}}).
-- returns: object_type, object_id, field, key, params (lihat docs/06-rules.md)
SELECT 'actions'::text, a.id, 'targetDate'::text, ''::text,
       jsonb_build_object('action', a.text, 'targetDate', to_char(a.target_date, 'DD Mon YYYY'),
                          'daysLate', (@today::date - a.target_date))
FROM actions a
WHERE a.package_id = @package_id
  AND a.status IN ('open', 'in_progress')
  AND a.target_date < @today::date
