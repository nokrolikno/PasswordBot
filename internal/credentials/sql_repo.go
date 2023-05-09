package credentials

import (
	"database/sql"
	"fmt"
	"sync"

	_ "github.com/go-sql-driver/mysql"
)

type CredentialsMemoryRepository struct {
	db *sql.DB
	mu *sync.Mutex
}

func NewMemoryRepo() *CredentialsMemoryRepository {
	dsn := "root:root@tcp(db:3306)/db?"
	dsn += "&charset=utf8"
	dsn += "&interpolateParams=true"

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		panic(err)
	}
	db.SetMaxOpenConns(10)

	err = db.Ping()
	if err != nil {
		panic(err)
	}
	db.Exec(
		`CREATE TABLE IF NOT EXISTS cred (
			UserID   int,
			Service  VARCHAR(50),
			Login    VARCHAR(100),
			Password VARCHAR(100)
		);`,
	)
	cmr := CredentialsMemoryRepository{
		db: db,
		mu: &sync.Mutex{},
	}
	return &cmr
}

func (repo *CredentialsMemoryRepository) Set(userID int64, service, login, password string) error {
	_, err := repo.db.Exec(
		"INSERT INTO cred (`UserID`, `Service`, `Login`, `Password`) VALUES (?, ?, ?, ?);",
		userID,
		service,
		login,
		password,
	)
	if err != nil {
		return fmt.Errorf("db error: %w", err)
	}
	return nil
}

func (repo *CredentialsMemoryRepository) Get(userID int64, service string) (string, string, bool, error) {
	var login, password string
	err := repo.db.
		QueryRow("SELECT Login, Password FROM cred WHERE UserID = ? AND Service = ?;", userID, service).
		Scan(&login, &password)
	if err != nil && err != sql.ErrNoRows {
		return "", "", false, fmt.Errorf("db error: %w", err)
	}
	if err == sql.ErrNoRows {
		return "", "", false, nil
	}
	return login, password, true, nil
}

func (repo *CredentialsMemoryRepository) Del(userID int64, service string) (bool, error) {
	result, err := repo.db.Exec(
		"DELETE FROM cred WHERE UserID = ? AND Service = ?;",
		userID,
		service,
	)
	if err != nil {
		return false, fmt.Errorf("db error: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("db error: %w", err)
	}
	if affected == 0 {
		return false, nil
	}
	return true, nil
}
