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
			Login    VARCHAR(50),
			Password VARCHAR(50)
		);`,
	)
	cmr := CredentialsMemoryRepository{
		db: db,
		mu: &sync.Mutex{},
	}
	return &cmr
}

func (repo *CredentialsMemoryRepository) Set(userID int64, service, login, password string) (bool, error) {
	var l, p string
	err := repo.db.
		QueryRow("SELECT Login, Password FROM cred WHERE UserID = ? AND Service = ?;", userID, service).
		Scan(&l, &p)
	if err != nil && err != sql.ErrNoRows {
		return false, fmt.Errorf("db error: %w", err)
	}
	if err == sql.ErrNoRows {
		_, err = repo.db.Exec(
			"INSERT INTO cred (`UserID`, `Service`, `Login`, `Password`) VALUES (?, ?, ?, ?);",
			userID,
			service,
			login,
			password,
		)
		if err != nil {
			return false, fmt.Errorf("db error: %w", err)
		}
		return true, nil
	}
	_, err = repo.db.Exec(
		"UPDATE cred SET Login = ?, Password = ? WHERE UserID = ? AND Service = ?;",
		login,
		password,
		userID,
		service,
	)
	if err != nil {
		return false, fmt.Errorf("db error: %w", err)
	}
	return false, nil
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

func (repo *CredentialsMemoryRepository) GetByID(userID int64) ([]string, error) {
	result := make([]string, 0)
	rows, err := repo.db.Query("SELECT Service FROM cred WHERE UserID = ?;", userID)
	if err != nil {
		return result, fmt.Errorf("db error: %w", err)
	}
	defer rows.Close()
	var service string
	for rows.Next() {
		err := rows.Scan(&service)
		if err != nil {
			return result, fmt.Errorf("db error: %w", err)
		}
		result = append(result, service)
	}
	return result, nil
}
