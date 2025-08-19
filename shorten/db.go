package shorten

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
)

type DataBaseConfig struct {
	Host	string
	Port	string
	User	string
	Password string
	DB	string
}

func (db *DataBaseConfig) Connect() (*pgx.Conn, error){
	conn, err := pgx.Connect(context.Background(), fmt.Sprintf("postgres://%s:%s@%s:%s/%s", db.User, db.Password, db.Host, db.Port, db.DB))
	if err != nil {
		return nil, err
	}
	return conn, nil
}