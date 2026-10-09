-- rule: S02
-- reads: characteristics, customer_sc_symbols
-- message: Symbol {{.symbol}} of characteristic {{.charNo}} has no entry in the symbol table of customer {{.customer}}.
-- returns: object_type, object_id, field, key, params (lihat docs/06-rules.md)
SELECT 'characteristics'::text, c.id, 'scSymbolId'::text, ''::text,
       jsonb_build_object('charNo', c.char_no, 'symbol', sc.code, 'customer', cu.name)
FROM characteristics c
JOIN sc_symbols sc ON sc.id = c.sc_symbol_id
JOIN packages pk ON pk.id = c.package_id
JOIN customers cu ON cu.id = pk.customer_id
WHERE c.package_id = @package_id
  AND NOT EXISTS (SELECT 1 FROM customer_sc_symbols m
                  WHERE m.customer_id = pk.customer_id AND m.sc_symbol_id = c.sc_symbol_id)
