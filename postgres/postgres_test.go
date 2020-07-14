package postgres

import "testing"

func TestConnectToPostgres(t *testing.T) {
	_, err := NewDB("", PostgresEnv{
		PostgresDatabase:   "fakedb",
		PostgresHost:       "fakehost",
		PostgresPassword:   "fakepassword",
		PostgresPort:       3000,
		PostgresSecureMode: false,
		PostgresUser:       "fakeuser",
	})

	if err == nil {
		t.Error("Passing wrong connection details doesn't result in an error")
	}
}
