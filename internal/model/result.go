package model

type MigrationResult struct {
	Row     int
	Email   string
	PSTFile string
	Status  string
	Error   string
}
