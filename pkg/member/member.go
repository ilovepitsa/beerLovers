package member

import (
	"database/sql"
	"time"
)

type Member struct {
	Id          int
	FIO         string
	Entry_Date  time.Time
	Address     sql.NullString
	PhoneNumber sql.NullString
	Email       sql.NullString
	Wallet_id   int
}
