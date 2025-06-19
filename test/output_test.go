package test

import (
	. "reconsiliation-service/imp"
	"testing"
)

func TestOutput_Reconciliation(t *testing.T) {
	if err := CreateRecords("../csv/st_small.csv", "systemTransaction", "20250601", "20250630"); err != nil {
		t.Errorf("Expected no error for invalid data type record, but got: %v", err)
	}

	if err := CreateRecords("../csv/bankA_20250605.csv", "bankStatement", "20250601", "20250630"); err != nil {
		t.Errorf("Expected no error for invalid data type record, but got: %v", err)
	}

	if err := CreateRecords("../csv/bankB_20250605.csv", "bankStatement", "20250601", "20250630"); err != nil {
		t.Errorf("Expected no error for invalid data type record, but got: %v", err)
	}

	output, err := Reconcile()
	if err != nil {
		t.Errorf("Expected no error during reconciliation, but got: %v", err)
	}

	if output == nil {
		t.Errorf("Expected output to be not nil, but got nil")
	}

	if output != nil && output.TotalProcessedRecords != 8 {
		t.Errorf("Expected TotalProcessedRecords to be 8, got: %d", output.TotalProcessedRecords)
	}

	if output.TotalMatchedTransactions != 4 {
		t.Errorf("Expected TotalMatchedTransactions to be 4, got: %d", output.TotalMatchedTransactions)
	}

	if output.TotalUnmatchedTransactions != 4 {
		t.Errorf("Expected TotalUnmatchedTransactions to be 4, got: %d", output.TotalUnmatchedTransactions)
	}

	if output.TotalInvalidRecords != 0 {
		t.Errorf("Expected TotalInvalidRecords to be 0, got: %d", output.TotalInvalidRecords)
	}

	if output.TotalDiscrepancies != 18796234.08 {
		t.Errorf("Expected TotalDiscrepancies to be 18796234.08, got: %.2f", output.TotalDiscrepancies)
	}

	if len(output.UnmatchedSystemTransactions) != 2 {
		t.Errorf("Expected UnmatchedSystemTransactions to be 2, got: %d", len(output.UnmatchedSystemTransactions))
	}

	for _, trx := range output.UnmatchedSystemTransactions {
		if trx.TrxID == "" || trx.Amount == 0 || trx.Type == "" || trx.TransactionTime == "" {
			t.Errorf("Unmatched transaction has empty fields: %+v", trx)
		}
	}

	if output.UnmatchedSystemTransactions[0].TrxID != "TX0011" {
		t.Errorf("Expected first unmatched transaction TrxID to be 'TX0011', got: %s", output.UnmatchedSystemTransactions[0].TrxID)
	}

	if output.UnmatchedSystemTransactions[0].Amount != 5943210.24 {
		t.Errorf("Expected first unmatched transaction Amount to be 5943210.24, got: %s", output.UnmatchedSystemTransactions[0].TrxID)
	}

	if output.UnmatchedBankStmts == nil {
		t.Errorf("Expected UnmatchedBankStmts to be not nil, but got nil")
	}

	if len(output.UnmatchedBankStmts) != 2 {
		t.Errorf("Expected UnmatchedBankStmts to have 2 banks, got: %d", len(output.UnmatchedBankStmts))
	}

	for bankName, stmts := range output.UnmatchedBankStmts {
		if bankName != "bankA" && bankName != "bankB" {
			t.Errorf("Expected bank name to be 'bankA' or 'bankB', got: %s", bankName)
		}

		if len(stmts) != 1 {
			t.Errorf("Expected each bank to have 1 unmatched statement, got: %d for bank %s", len(stmts), bankName)
		}

		for _, stmt := range stmts {
			if stmt.UniqueIdentifier == "" || stmt.Amount == 0 || stmt.Date == "" {
				t.Errorf("Unmatched bank statement has empty fields: %+v", stmt)
			}
		}
	}

	if output.UnmatchedBankStmts["bankA"][0].UniqueIdentifier != "BA0003" {
		t.Errorf("Expected unmatched bank statement UniqueIdentifier to be 'BA0003', got: %s", output.UnmatchedBankStmts["bankA"][0].UniqueIdentifier)
	}

	if output.UnmatchedBankStmts["bankA"][0].Amount != 3955387.85 {
		t.Errorf("Expected unmatched bank statement Amount to be 3955387.85, got: %0.2f", output.UnmatchedBankStmts["bankA"][0].Amount)
	}

	if output.UnmatchedBankStmts["bankB"][0].UniqueIdentifier != "BA0048" {
		t.Errorf("Expected unmatched bank statement UniqueIdentifier to be 'BA0048', got: %s", output.UnmatchedBankStmts["bankB"][0].UniqueIdentifier)
	}

	if output.UnmatchedBankStmts["bankB"][0].Amount != 5405730.98 {
		t.Errorf("Expected unmatched bank statement Amount to be 5405730.98, got: %0.2f", output.UnmatchedBankStmts["bankB"][0].Amount)
	}

	if output.UnmatchedBankStmts["bankA"][0].Date != "2025-06-05" {
		t.Errorf("Expected unmatched bank statement Date to be '2025-06-05', got: %s", output.UnmatchedBankStmts["bankA"][0].Date)
	}

	if output.UnmatchedBankStmts["bankB"][0].Date != "2025-06-05" {
		t.Errorf("Expected unmatched bank statement Date to be '2025-06-05', got: %s", output.UnmatchedBankStmts["bankB"][0].Date)
	}
}
