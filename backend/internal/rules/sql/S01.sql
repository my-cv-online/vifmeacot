-- rule: S01
-- reads: characteristics, failure_modes, cp_lines
-- message: Special characteristic {{.charNo}} {{.charName}} ({{.symbol}}) does not appear in the {{.missingIn}}.
-- returns: object_type, object_id, field, key, params (lihat docs/06-rules.md)
SELECT 'characteristics'::text, c.id, 'scSymbolId'::text, x.missing_in::text,
       jsonb_build_object('charNo', c.char_no, 'charName', c.name, 'symbol', sc.code, 'missingIn', x.missing_in)
FROM characteristics c
JOIN sc_symbols sc ON sc.id = c.sc_symbol_id
CROSS JOIN LATERAL (
  SELECT 'PFMEA' AS missing_in
  WHERE NOT EXISTS (SELECT 1 FROM failure_modes fm WHERE fm.characteristic_id = c.id)
  UNION ALL
  SELECT 'Control Plan'
  WHERE NOT EXISTS (SELECT 1 FROM cp_lines l WHERE l.characteristic_id = c.id)
) x
WHERE c.package_id = @package_id
