UPDATE characters
SET data_json = json_set(
    json_remove(data_json, '$.stats'),
    '$.abilities',
    json_object(
        'Strength', json_object('current', COALESCE(CAST(json_extract(data_json, '$.stats.strength') AS INTEGER), 0), 'max', COALESCE(CAST(json_extract(data_json, '$.stats.strength') AS INTEGER), 0)),
        'Agility', json_object('current', COALESCE(CAST(json_extract(data_json, '$.stats.agility') AS INTEGER), 0), 'max', COALESCE(CAST(json_extract(data_json, '$.stats.agility') AS INTEGER), 0)),
        'Presence', json_object('current', COALESCE(CAST(json_extract(data_json, '$.stats.presence') AS INTEGER), 0), 'max', COALESCE(CAST(json_extract(data_json, '$.stats.presence') AS INTEGER), 0)),
        'Toughness', json_object('current', COALESCE(CAST(json_extract(data_json, '$.stats.toughness') AS INTEGER), 0), 'max', COALESCE(CAST(json_extract(data_json, '$.stats.toughness') AS INTEGER), 0)),
        'Knowledge', json_object('current', COALESCE(CAST(json_extract(data_json, '$.stats.knowledge') AS INTEGER), 0), 'max', COALESCE(CAST(json_extract(data_json, '$.stats.knowledge') AS INTEGER), 0))
    )
)
WHERE json_extract(data_json, '$.stats') IS NOT NULL;
