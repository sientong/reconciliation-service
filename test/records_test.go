package test

import (
	"testing"

	. "github.com/sientong/reconciliation-service/imp"
	"github.com/sientong/reconciliation-service/model"
)

func TestRecrod_WithValidTransactionRecord(t *testing.T) {

	clearRecords()
	err := CreateRecords("../csv/st_small.csv", "systemTransaction", "20250601", "20250630")
	if err != nil {
		t.Errorf("Expected no error for valid record, but got: %v", err)
	}

	if len(model.SystemTransactionRecords) != 6 {
		t.Errorf("Expected 6 system transaction records, got: %d", len(model.SystemTransactionRecords))
	}

	for _, record := range model.SystemTransactionRecords {
		if record.TrxID == "" || record.Amount == 0 || record.Type == "" || record.TransactionTime == "" {
			t.Errorf("Record has empty fields: %+v", record)
		}
	}

	if model.SystemTransactionRecords[0].TrxID != "TX0001" {
		t.Errorf("Expected first record TrxID to be 'TX0001', got: %s", model.SystemTransactionRecords[0].TrxID)
	}

	if model.SystemTransactionRecords[0].Amount != 6241250.16 {
		t.Errorf("Expected first record Amount to be 6241250.16, got: %s", model.SystemTransactionRecords[0].TrxID)
	}

	if model.SystemTransactionRecords[0].Type != "DEBIT" {
		t.Errorf("Expected first record Type to be 'DEBIT', got: %s", model.SystemTransactionRecords[0].TrxID)
	}

	if model.SystemTransactionRecords[0].TransactionTime != "2025-06-05T08:01:00Z" {
		t.Errorf("Expected first record Type to be '2025-06-05T08:01:00Z', got: %s", model.SystemTransactionRecords[0].TrxID)
	}

	if model.SystemTransactionRecords[0].IsMatched {
		t.Errorf("Expected first record IsMatched to be false, got: %t", model.SystemTransactionRecords[0].IsMatched)
	}

	clearRecords()
}

func TestRecord_WithValidBankStatementRecord(t *testing.T) {

	clearRecords()

	err := CreateRecords("../csv/bankA_20250605.csv", "bankStatement", "20250601", "20250630")
	if err != nil {
		t.Errorf("Expected no error for valid record, but got: %v", err)
	}

	for bankName := range model.BankStatementRecordsMap {
		if bankName != "bankA" {
			t.Errorf("Expected bank name to be 'bankA', got: %s", bankName)
		}
	}

	if len(model.BankStatementRecordsMap["bankA"]) != 3 {
		t.Errorf("Expected 2 bank statement records, got: %d", len(model.BankStatementRecordsMap["bankA"]))
	}

	for _, record := range model.BankStatementRecordsMap["bankA"] {
		if record.UniqueIdentifier == "" || record.Amount == 0 || record.Date == "" {
			t.Errorf("Record has empty fields: %+v", record)
		}
	}

	if model.BankStatementRecordsMap["bankA"][0].UniqueIdentifier != "BA0001" {
		t.Errorf("Expected first record UniqueIndentifier to be 'BA0001', got: %s", model.SystemTransactionRecords[0].TrxID)
	}

	if model.BankStatementRecordsMap["bankA"][0].Amount != -6241250.16 {
		t.Errorf("Expected first record Amount to be -6241250.16, got: %s", model.SystemTransactionRecords[0].TrxID)
	}

	if model.BankStatementRecordsMap["bankA"][0].Date != "2025-06-05" {
		t.Errorf("Expected first record Type to be '2025-06-05', got: %s", model.SystemTransactionRecords[0].TrxID)
	}

	if model.BankStatementRecordsMap["bankA"][0].IsMatched {
		t.Errorf("Expected first record IsMatched to be false, got: %t", model.SystemTransactionRecords[0].IsMatched)
	}

	clearRecords()
}

func TestRecord_WithIncorrectDataType(t *testing.T) {

	clearRecords()

	err := CreateRecords("../csv/st_incorrect_record.csv", "systemTransaction", "20250601", "20250630")
	if err != nil {
		t.Errorf("Expected no error for invalid data type record, but got: %v", err)
	}

	if len(model.SystemTransactionRecords) != 0 {
		t.Errorf("Expected no system transaction records, got: %d", len(model.SystemTransactionRecords))
	}

	clearRecords()
}

func clearRecords() {
	model.SystemTransactionRecords = nil
	model.BankStatementRecordsMap = make(map[string][]*model.BankStatementRecord)
}
