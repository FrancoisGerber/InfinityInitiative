package utils

import (
	"os"
)

// Secret when creating sessions
func Secret() []byte {
	return []byte(os.Getenv("SECRET"))
}

// Local SQLite DB name
func DBName() string {
	return os.Getenv("DBNAME")
}

// Connection to Turso
func Token() string {
	return os.Getenv("REMOTEDBTOKEN")
}

func Remote() string {
	return os.Getenv("REMOTEDB")
}

func RemoteConn() string {
	return Remote() + "?authToken=" + Token()
}
