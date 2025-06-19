package model

type BankStatementRecord struct {
	UniqueIdentifier string
	Amount           float64
	Date             string
	IsMatched        bool
}

var BankStatementRecordsMap = make(map[string][]BankStatementRecord)
