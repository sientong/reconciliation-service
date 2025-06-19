package model

type InternalTransactionRecord struct {
	TrxID           string
	Amount          float64
	Type            string
	TransactionTime string
	IsMatched       bool
}

var SystemTransactionRecords []InternalTransactionRecord
