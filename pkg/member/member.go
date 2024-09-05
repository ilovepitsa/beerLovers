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
	Email       string
	Wallet_id   int
	Balance     float32
	User_level  string `sql:"type:user_level"`
}

// ₽
func (m *Member) GetID() int {
	return m.Id
}

func (m *Member) IsAdmin() bool {
	return m.User_level == "admin"
}
