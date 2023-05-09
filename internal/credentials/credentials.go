package credentials

type Credentials struct {
	userID   int64
	service  string
	login    string
	password string
}

type CredentialsRepo interface {
	Set(userID int64, service, login, password string) (bool, error)
	Get(userID int64, service string) (string, string, bool, error)
	Del(userID int64, service string) (bool, error)
	GetByID(userID int64) ([]string, error)
}
