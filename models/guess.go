package models

import (
	"context"
	"time"

	"github.com/MixinNetwork/supergroup/session"
	"github.com/jackc/pgx/v4"
)

const guess_DDL = `
CREATE TABLE IF NOT EXISTS guess (
	client_id VARCHAR(36) NOT NULL,
	guess_id  VARCHAR(36) NOT NULL,
	symbol 	  VARCHAR NOT NULL,
	price_usd VARCHAR NOT NULL,
	rules 	 VARCHAR NOT NULL,
	explain VARCHAR NOT NULL,
	start_time VARCHAR NOT NULL,
	end_time VARCHAR NOT NULL,
	start_at TIMESTAMP NOT NULL,
	end_at  TIMESTAMP NOT NULL,
	created_at TIMESTAMP NOT NULL DEFAULT now()
);
`

const guess_record_DDL = `
CREATE TABLE IF NOT EXISTS guess_record (
	guess_id  VARCHAR(36) NOT NULL,
	user_id VARCHAR(36) NOT NULL,
	guess_type SMALLINT NOT NULL,
	date DATE NOT NULL,
	result SMALLINT NOT NULL DEFAULT 0,
	PRIMARY KEY (guess_id, user_id, date)
);
`

type Guess struct {
	ClientId  string    `json:"client_id"`
	GuessId   string    `json:"guess_id"`
	Symbol    string    `json:"symbol"`
	PriceUsd  string    `json:"price_usd"`
	Rules     string    `json:"rules"`
	Explain   string    `json:"explain"`
	StartTime string    `json:"start_time"`
	EndTime   string    `json:"end_time"`
	StartAt   time.Time `json:"start_at"`
	EndAt     time.Time `json:"end_at"`
	CreatedAt time.Time `json:"created_at"`
}

type GuessRecord struct {
	GuessId   string `json:"guess_id"`
	UserId    string `json:"user_id"`
	GuessType int    `json:"guess_type"`
	Result    int    `json:"result"`
	Date      string `json:"date"`
}

const (
	GuessTypeUP = iota + 1
	GuessTypeDown
	GuessTypeFlat

	GuessResultPending = iota
	GuessResultWin
	GuessResultLose
	GuessResultDraw
)

type GuessPageResp struct {
	*Guess
	TodayGuess int `json:"today_guess"`
}

func GetGuessPageInitData(ctx context.Context, u *ClientUser, guessID string) (*GuessPageResp, error) {
	guess, err := getGuessByID(ctx, guessID)
	if err != nil {
		return nil, err
	}
	todyGuessType, _ := getTodayGuessType(ctx, u.UserID, guessID)
	resp := &GuessPageResp{
		Guess:      guess,
		TodayGuess: todyGuessType,
	}
	return resp, nil
}

func PostGuess(ctx context.Context, u *ClientUser, guessID string, guessType int) error {
	_, err := session.Database(ctx).Exec(ctx, `
INSERT INTO guess_record (guess_id, user_id, guess_type, date, result)
VALUES ($1, $2, $3, now(), $4)`,
		guessID, u.UserID, guessType, GuessResultPending)
	return err
}

func GetGuessRecordListByUserID(ctx context.Context, u *ClientUser, guessID string, page int) ([]*GuessRecord, error) {
	if page < 1 {
		page = 1
	}
	return getGuessRecordByUserID(ctx, u.UserID, guessID, page)
}

func getGuessByID(ctx context.Context, guessID string) (*Guess, error) {
	var g Guess
	err := session.Database(ctx).QueryRow(ctx, `
SELECT client_id, guess_id, symbol, price_usd, rules, explain, start_time, end_time, start_at, end_at, created_at 
FROM guess 
WHERE guess_id = $1`, guessID).
		Scan(&g.ClientId, &g.GuessId, &g.Symbol, &g.PriceUsd, &g.Rules, &g.Explain, &g.StartTime, &g.EndTime, &g.StartAt, &g.EndAt, &g.CreatedAt)
	return &g, err
}

func getTodayGuessType(ctx context.Context, userID string, guessID string) (int, error) {
	var guessType int
	err := session.Database(ctx).QueryRow(ctx, `
SELECT guess_type 
FROM guess_record 
WHERE user_id = $1 AND guess_id = $2 AND date=now()`,
		userID, guessID).Scan(&guessType)
	return guessType, err
}

func getGuessRecordByUserID(ctx context.Context, userID string, guessID string, page int) ([]*GuessRecord, error) {
	grs := make([]*GuessRecord, 0)
	err := session.Database(ctx).ConnQuery(ctx, `
SELECT guess_id, user_id, guess_type, date, result 
FROM guess_record 
WHERE user_id = $1 AND guess_id = $2
ORDER BY date DESC
OFFSET $3 LIMIT 10
`, func(rows pgx.Rows) error {
		for rows.Next() {
			var g GuessRecord
			if err := rows.Scan(&g.GuessId, &g.UserId, &g.GuessType, &g.Date, &g.Result); err != nil {
				return err
			}
			grs = append(grs, &g)
		}
		return nil
	}, userID, guessID, (page-1)*10)
	return grs, err
}

func timeToUpdateGuessResult(ctx context.Context) {

}
