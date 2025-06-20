package model

type Output struct {
	TotalProcessedRecords            int
	TotalMatchedTransactions         int
	TotalUnmatchedTransactions       int
	TotalUnmatchedSystemTransactions int
	TotalUnmatchedBankStmts          int
	TotalInvalidRecords              int
	TotalDiscrepancies               float64
	UnmatchedSystemTransactions      []InternalTransactionRecord
	UnmatchedBankStmts               map[string][]BankStatementRecord
}
