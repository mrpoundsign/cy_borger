UPDATE characters
SET data_json = json_set(
    json_remove(data_json, '$.abilities'),
    '$.stats',
    json_object(
        'strength', COALESCE(CAST(json_extract(data_json, '$.abilities.Strength.current') AS INTEGER), CAST(json_extract(data_json, '$.abilities.Strength') AS INTEGER), CAST(json_extract(data_json, '$.abilities.strength') AS INTEGER), 0),
        'agility', COALESCE(CAST(json_extract(data_json, '$.abilities.Agility.current') AS INTEGER), CAST(json_extract(data_json, '$.abilities.Agility') AS INTEGER), CAST(json_extract(data_json, '$.abilities.agility') AS INTEGER), 0),
        'presence', COALESCE(CAST(json_extract(data_json, '$.abilities.Presence.current') AS INTEGER), CAST(json_extract(data_json, '$.abilities.Presence') AS INTEGER), CAST(json_extract(data_json, '$.abilities.presence') AS INTEGER), 0),
        'toughness', COALESCE(CAST(json_extract(data_json, '$.abilities.Toughness.current') AS INTEGER), CAST(json_extract(data_json, '$.abilities.Toughness') AS INTEGER), CAST(json_extract(data_json, '$.abilities.toughness') AS INTEGER), 0),
        'knowledge', COALESCE(CAST(json_extract(data_json, '$.abilities.Knowledge.current') AS INTEGER), CAST(json_extract(data_json, '$.abilities.Knowledge') AS INTEGER), CAST(json_extract(data_json, '$.abilities.knowledge') AS INTEGER), 0)
    )
)
WHERE json_extract(data_json, '$.abilities') IS NOT NULL;
