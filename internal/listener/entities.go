package listener

import (
	"database/sql"
	"time"
)

type Alert struct {
	ID             int64        `db:"id"`
	Email          string       `db:"email"`
	Symbol         string       `db:"symbol"`
	Price          int64        `db:"price"`
	MarkedToSendAt sql.NullTime `db:"marked_to_send_at"`
	SendAt         sql.NullTime `db:"sent_at"`
	CreatedAt      time.Time    `db:"created_at"`
	UpdatedAt      time.Time    `db:"updated_at"`
}
