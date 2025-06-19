package model

type Output struct {
	TotalProcessedRecords       int
	TotalMatchedTransactions    int
	TotalUnmatchedTransactions  int
	UnmatchedSystemTransactions []InternalTransactionRecord
	UnmatchedBankStmts          []BankStatementRecord
	TotalInvalidRecords         int
	TotalDiscrepancies          float64
}
