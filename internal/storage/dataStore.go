package storage

import "github.com/ethannself/cloud-drive-b/internal/database"

var dataStore *database.DataStore

func InitDataStore() {
	dataStore = database.New()
}

func GetDataStore() *database.DataStore {
	return dataStore
}
