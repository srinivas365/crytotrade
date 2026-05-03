package db

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type User struct {
	ID           string    `json:"id"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"password_hash"`
	CreatedAt    time.Time `json:"created_at"`
}

type UserSettings struct {
	UserID           string  `json:"user_id"`
	ThresholdPct     float64 `json:"threshold_pct"`
	TelegramBotToken string  `json:"telegram_bot_token"`
	TelegramChatID   string  `json:"telegram_chat_id"`
	InAppAlerts      bool    `json:"in_app_alerts"`
	AlertSound       bool    `json:"alert_sound"`
}

type AlertRecord struct {
	ID           string    `json:"id"`
	UserID       string    `json:"user_id"`
	Symbol       string    `json:"symbol"`
	BuyExchange  string    `json:"buy_exchange"`
	SellExchange string    `json:"sell_exchange"`
	SpreadPct    float64   `json:"spread_pct"`
	BuyPrice     float64   `json:"buy_price"`
	SellPrice    float64   `json:"sell_price"`
	FiredAt      time.Time `json:"fired_at"`
}

type Queries struct{ pool *pgxpool.Pool }

func NewQueries(pool *pgxpool.Pool) *Queries { return &Queries{pool: pool} }

func (q *Queries) CreateUser(ctx context.Context, email, hash string) (*User, error) {
	u := &User{}
	if err := q.pool.QueryRow(ctx,
		`INSERT INTO users(email,password_hash) VALUES($1,$2)
		 RETURNING id,email,password_hash,created_at`,
		email, hash,
	).Scan(&u.ID, &u.Email, &u.PasswordHash, &u.CreatedAt); err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}
	if _, err := q.pool.Exec(ctx,
		`INSERT INTO user_settings(user_id) VALUES($1) ON CONFLICT(user_id) DO NOTHING`, u.ID,
	); err != nil {
		return nil, fmt.Errorf("create settings: %w", err)
	}
	return u, nil
}

func (q *Queries) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	u := &User{}
	if err := q.pool.QueryRow(ctx,
		`SELECT id,email,password_hash,created_at FROM users WHERE email=$1`, email,
	).Scan(&u.ID, &u.Email, &u.PasswordHash, &u.CreatedAt); err != nil {
		return nil, fmt.Errorf("get user: %w", err)
	}
	return u, nil
}

func (q *Queries) GetSettings(ctx context.Context, userID string) (*UserSettings, error) {
	s := &UserSettings{}
	if err := q.pool.QueryRow(ctx,
		`SELECT user_id,threshold_pct,telegram_bot_token,telegram_chat_id,in_app_alerts,alert_sound
		 FROM user_settings WHERE user_id=$1`, userID,
	).Scan(&s.UserID, &s.ThresholdPct, &s.TelegramBotToken, &s.TelegramChatID, &s.InAppAlerts, &s.AlertSound); err != nil {
		return nil, fmt.Errorf("get settings for user %s: %w", userID, err)
	}
	return s, nil
}

func (q *Queries) UpsertSettings(ctx context.Context, s *UserSettings) error {
	_, err := q.pool.Exec(ctx,
		`INSERT INTO user_settings(user_id,threshold_pct,telegram_bot_token,telegram_chat_id,in_app_alerts,alert_sound)
		 VALUES($1,$2,$3,$4,$5,$6)
		 ON CONFLICT(user_id) DO UPDATE SET
		   threshold_pct=EXCLUDED.threshold_pct,
		   telegram_bot_token=EXCLUDED.telegram_bot_token,
		   telegram_chat_id=EXCLUDED.telegram_chat_id,
		   in_app_alerts=EXCLUDED.in_app_alerts,
		   alert_sound=EXCLUDED.alert_sound`,
		s.UserID, s.ThresholdPct, s.TelegramBotToken, s.TelegramChatID, s.InAppAlerts, s.AlertSound,
	)
	return err
}

func (q *Queries) GetAllSettings(ctx context.Context) ([]*UserSettings, error) {
	rows, err := q.pool.Query(ctx,
		`SELECT user_id,threshold_pct,telegram_bot_token,telegram_chat_id,in_app_alerts,alert_sound FROM user_settings`)
	if err != nil {
		return nil, fmt.Errorf("get all settings: %w", err)
	}
	defer rows.Close()
	var out []*UserSettings
	for rows.Next() {
		s := &UserSettings{}
		if err := rows.Scan(&s.UserID, &s.ThresholdPct, &s.TelegramBotToken, &s.TelegramChatID, &s.InAppAlerts, &s.AlertSound); err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	return out, rows.Err()
}

func (q *Queries) InsertAlert(ctx context.Context, r *AlertRecord) error {
	_, err := q.pool.Exec(ctx,
		`INSERT INTO alert_history(user_id,symbol,buy_exchange,sell_exchange,spread_pct,buy_price,sell_price)
		 VALUES($1,$2,$3,$4,$5,$6,$7)`,
		r.UserID, r.Symbol, r.BuyExchange, r.SellExchange, r.SpreadPct, r.BuyPrice, r.SellPrice,
	)
	return err
}

func (q *Queries) CountAlertHistory(ctx context.Context, userID string) (int, error) {
	var n int
	err := q.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM alert_history WHERE user_id=$1`, userID,
	).Scan(&n)
	return n, err
}

func (q *Queries) GetAlertHistory(ctx context.Context, userID string, limit, offset int) ([]*AlertRecord, error) {
	rows, err := q.pool.Query(ctx,
		`SELECT id,user_id,symbol,buy_exchange,sell_exchange,spread_pct,buy_price,sell_price,fired_at
		 FROM alert_history WHERE user_id=$1 ORDER BY fired_at DESC LIMIT $2 OFFSET $3`,
		userID, limit, offset,
	)
	if err != nil {
		return nil, fmt.Errorf("get alert history for user %s: %w", userID, err)
	}
	defer rows.Close()
	var out []*AlertRecord
	for rows.Next() {
		r := &AlertRecord{}
		if err := rows.Scan(&r.ID, &r.UserID, &r.Symbol, &r.BuyExchange, &r.SellExchange,
			&r.SpreadPct, &r.BuyPrice, &r.SellPrice, &r.FiredAt); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}
