-- rule: S04
-- reads: characteristics, failure_modes, failure_effects
-- message: Characteristic {{.charNo}} is classified {{.symbol}}, but its highest S {{.maxS}} is below the customer requirement ({{.minS}}).
-- returns: object_type, object_id, field, key, params (lihat docs/06-rules.md)
SELECT 'characteristics'::text, c.id, 'scSymbolId'::text, ''::text,
       jsonb_build_object('charNo', c.char_no, 'symbol', sc.code, 'maxS', f.max_s, 'minS', cu.cc_min_severity)
FROM characteristics c
JOIN sc_symbols sc ON sc.id = c.sc_symbol_id AND sc.is_critical
JOIN packages pk ON pk.id = c.package_id
JOIN customers cu ON cu.id = pk.customer_id AND cu.cc_min_severity IS NOT NULL
JOIN LATERAL (
  SELECT max(e.s) AS max_s
  FROM failure_modes fm JOIN failure_effects e ON e.failure_mode_id = fm.id
  WHERE fm.characteristic_id = c.id
) f ON f.max_s IS NOT NULL                     -- karakteristik tanpa failure mode dilaporkan oleh S01
WHERE c.package_id = @package_id
  AND f.max_s < cu.cc_min_severity
