package db

import (
	"strings"
	"time"
)

type GameLog struct {
	ID          int       `json:"id"`
	GameID      string    `json:"game_id"`
	CharacterID string    `json:"character_id"`
	EventType   string    `json:"event_type"`
	Message     string    `json:"message"`
	CreatedAt   time.Time `json:"created_at"`
}

func (d *DB) LogGameEvent(gameID, characterID, eventType, message string) error {
	query := `INSERT INTO game_logs (game_id, character_id, event_type, message) VALUES (?, ?, ?, ?)`
	_, err := d.conn.Exec(query, gameID, characterID, eventType, message)
	return err
}

func (d *DB) GetGameLogs(gameID string, types []string, limit int) ([]GameLog, error) {
	query := `SELECT id, game_id, character_id, event_type, message, created_at FROM game_logs WHERE game_id = ?`
	args := []interface{}{gameID}

	if len(types) > 0 {
		placeholders := make([]string, len(types))
		for i, t := range types {
			placeholders[i] = "?"
			args = append(args, t)
		}
		query += ` AND event_type IN (` + strings.Join(placeholders, ",") + `)`
	}

	query += ` ORDER BY id DESC LIMIT ?`
	args = append(args, limit)

	rows, err := d.conn.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []GameLog
	for rows.Next() {
		var l GameLog
		if err := rows.Scan(&l.ID, &l.GameID, &l.CharacterID, &l.EventType, &l.Message, &l.CreatedAt); err != nil {
			return nil, err
		}
		logs = append(logs, l)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return logs, nil
}
