package shorten

import (
	"context"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5"
)

type DataBaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DB       string
}

type Table struct {
	Id    int
	Slug  string
	OgUrl string
}

func (db *DataBaseConfig) Connect() (*pgx.Conn, error) {
	conn, err := pgx.Connect(context.Background(), fmt.Sprintf("postgresql://%s:%s@%s:%s/%s", db.User, db.Password, db.Host, db.Port, db.DB))
	if err != nil {
		return nil, err
	}
	return conn, nil
}

func CreateTable(conn *pgx.Conn) error {
	_, err := conn.Exec(context.Background(), `
		CREATE TABLE IF NOT EXISTS urls (
		id SERIAL PRIMARY KEY,
		slug VARCHAR(3),
		ogurl TEXT
		);
	`)
	if err != nil {
		return err
	}
	return nil
}

func ReadEntry(conn *pgx.Conn) *[]Table {
	rows, err := conn.Query(context.Background(), `
		SELECT * FROM urls;
	`)
	if err != nil {
		err = CreateTable(conn)
		if err != nil {
			log.Println("Failed to Create Table: ", err)
		}
		return ReadEntry(conn)
	}
	defer rows.Close()
	var table []Table
	for rows.Next() {
		var t Table
		err = rows.Scan(&t.Id, &t.Slug, &t.OgUrl)
		if err != nil {
			log.Println("Scan failed: ", err)
		}
		table = append(table, t)
	}
	return &table
}

func CreateEntry(slug string, ogurl string, conn *pgx.Conn) error {
	_, err := conn.Exec(context.Background(), fmt.Sprintf(`INSERT INTO urls (slug, ogurl) VALUES ('%s', '%s');`, slug, ogurl))
	if err != nil {
		return err
	}
	return nil
}

func PopulateMap(us *URLShortener) {
	conn, err := us.DbConfig.Connect()
	if err != nil {
		log.Println("Failed to connect to DB: ", err)
	}
	defer conn.Close(context.Background())
	table := ReadEntry(conn)
	for _, t := range *table {
		us.Urls[t.Slug] = t.OgUrl
	}
}
