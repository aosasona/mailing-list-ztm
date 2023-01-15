package db

import (
	"database/sql"
	"log"
	"time"

	"github.com/mattn/go-sqlite3"
)

type EmailEntry struct {
	Id          int64
	Email       string
	ConfirmedAt *time.Time
	OptOut      bool
}

func TryCreate(db *sql.DB) {
	_, err := db.Exec(`
  CREATE TABLE IF NOT EXISTS emails (
    id 							INTEGER PRIMARY KEY,
    email 					TEXT UNIQUE,
    confirmed_at 		INTEGER,
    opt_out 				INTEGER
  );`)
	if err != nil {
		if sqlErr, ok := err.(sqlite3.Error); ok {
			if sqlErr.Code != 1 {
				log.Fatal(sqlErr)
			}
		} else {
			log.Fatal(err)
		}
	}
}

func emailEntryFromRow(row *sql.Rows) (*EmailEntry, error) {
	var (
		id          int64
		email       string
		confirmedAt int64
		optOut      bool
	)

	err := row.Scan(&id, &email, &confirmedAt, &optOut)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	t := time.Unix(confirmedAt, 0)

	return &EmailEntry{
		Id:          id,
		Email:       email,
		ConfirmedAt: &t,
		OptOut:      optOut,
	}, nil
}

func CreateEmail(db *sql.DB, email string) error {
	_, err := db.Exec(`INSERT INTO emails(email, confirmed_at, opt_out) VALUES(?, 0, false)`, email)
	if err != nil {
		log.Println(err)
		return err
	}

	return nil
}

func GetEmail(db *sql.DB, email string) (*EmailEntry, error) {
	rows, err := db.Query(
		`SELECT id, email, confirmed_at, opt_out FROM emails WHERE email = ?`,
		email,
	)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	defer rows.Close()

	for rows.Next() {
		return emailEntryFromRow(rows)
	}

	return nil, nil
}

func UpdateEmail(db *sql.DB, entry EmailEntry) error {
	t := entry.ConfirmedAt.Unix()

	_, err := db.Exec(
		`INSERT INTO emails(email, confirmed_at, opt_out) VALUES(?,?,?) ON CONFLICT(email) DO UPDATE SET confirmed_at = ?, opt_out = ?`,
		entry.Email,
		t,
		entry.OptOut,
		t,
		entry.OptOut,
	)
	if err != nil {
		log.Println(err)
		return err
	}

	return nil
}

func DeleteEmail(db *sql.DB, email string) error {
	_, err := db.Exec(`UPDATE emails SET opt_out = true WHERE email = ?`, email)
	if err != nil {
		log.Println(err)
		return err
	}

	return nil
}

type GetEmailsParams struct {
	Page  int
	Count int
}

func GetEmails(db *sql.DB, params GetEmailsParams) ([]EmailEntry, error) {
	var e []EmailEntry

	offset := (params.Page - 1) * params.Count

	rows, err := db.Query(
		`SELECT id, email, confirmed_at, opt_out FROM emails WHERE opt_out = false ORDER by id ASC LIMIT ? OFFSET ?`,
		params.Count,
		offset,
	)
	if err != nil {
		log.Println(err)
		return e, err
	}

	defer rows.Close()

	emails := make([]EmailEntry, 0, params.Count)

	for rows.Next() {
		email, err := emailEntryFromRow(rows)
		if err != nil {
			return nil, err
		}

		emails = append(emails, *email)
	}

	return emails, nil
}
