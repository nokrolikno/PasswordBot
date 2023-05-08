package credentials

import (
	"sync"
)

type CredentialsMemoryRepository struct {
	data []*Credentials
	mu   *sync.Mutex
}

func NewMemoryRepo() *CredentialsMemoryRepository {
	cmr := CredentialsMemoryRepository{
		data: make([]*Credentials, 0, 10),
		mu:   &sync.Mutex{},
	}
	return &cmr
}

func (repo *CredentialsMemoryRepository) Set(userID int64, service, login, password string) error {
	cred := &Credentials{userID: userID, service: service, login: login, password: password}
	for _, val := range repo.data {
		if *cred == *val {
			return nil
		}
	}
	repo.data = append(repo.data, cred)
	return nil
}

func (repo *CredentialsMemoryRepository) Get(userID int64, service string) (string, string, bool, error) {
	for _, val := range repo.data {
		if val.userID != userID || val.service != service {
			continue
		}
		return val.login, val.password, true, nil
	}
	return "", "", false, nil
}

func (repo *CredentialsMemoryRepository) Del(userID int64, service string) (bool, error) {
	for idx, val := range repo.data {
		if val.userID != userID || val.service != service {
			continue
		}
		repo.data = append(repo.data[:idx], repo.data[idx+1:]...)
		return true, nil
	}
	return false, nil
}
