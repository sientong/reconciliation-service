package validator

import (
	"fmt"
	"time"
)

func ValidateRecord(record []string, recordType string) error {

	switch recordType {
	case "systemTransaction":
		validateSystemTransactionRecord(record)
	case "bankStatement":
		validateBankStatementRecord(record)
	default:
		return fmt.Errorf("unknown record type: %s", recordType)
	}

	return nil
}

func validateSystemTransactionRecord(record []string) error {
	if len(record) != 4 {
		return fmt.Errorf("expected 4 columns, got %d", len(record))
	}

	if record[0] == "" || record[1] == "" || record[2] == "" || record[3] == "" {
		return fmt.Errorf("all columns must be filled in for system transaction record")
	}

	if record[2] != "credit" && record[2] != "debit" {
		return fmt.Errorf("invalid transaction type: %s, expected 'credit' or 'debit'", record[2])
	}

	if _, err := time.Parse("20060102", record[3]); err != nil {
		return fmt.Errorf("invalid transaction time format: %v, expected YYYYMMDD", err)
	}

	return nil
}

func validateBankStatementRecord(record []string) error {
	if len(record) != 3 {
		return fmt.Errorf("expected 3 columns, got %d", len(record))
	}

	if record[0] == "" || record[1] == "" || record[2] == "" {
		return fmt.Errorf("all columns must be filled in for bank statement record")
	}

	if _, err := time.Parse("20060102", record[2]); err != nil {
		return fmt.Errorf("invalid date format: %v, expected YYYYMMDD", err)
	}

	return nil
}
