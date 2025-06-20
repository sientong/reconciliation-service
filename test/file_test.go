package test

import (
	"fmt"
	. "reconciliation-service/validator"
	"testing"
)

func TestFile_WithNotExistFile(t *testing.T) {
	filepath := "non_existent_file.csv"
	err := ValidateFile(filepath, "systemTransaction")
	if err == nil {
		t.Errorf("Expected error for non-existent file, but got nil")
	}
	expectedMessage := "open non_existent_file.csv: no such file or directory"
	if err.Error() != expectedMessage {
		t.Errorf("Expected error message '%s', but got '%s'", expectedMessage, err.Error())
	}
}

func TestFile_WithIncorrectNumberHeaderColumn(t *testing.T) {
	filepath := "../csv/st_incorrect_header_column.csv"
	err := ValidateFile(filepath, "systemTransaction")
	if err == nil {
		t.Errorf("Expected error for incorrect header, but got nil")
	}

	expectedMessage := fmt.Sprintf("invalid header in %s: expected 4 columns, got 3", filepath)
	if err.Error() != expectedMessage {
		t.Errorf("Expected error message '%s', but got '%s'", expectedMessage, err.Error())
	}
}

func TestFile_WithIncorrectHeaderValue(t *testing.T) {
	filepath := "../csv/st_incorrect_header_value.csv"
	err := ValidateFile(filepath, "systemTransaction")
	if err == nil {
		t.Errorf("Expected error for incorrect header, but got nil")
	}

	expectedMessage := fmt.Sprintf("invalid header in %s: expected column trxID, got transactionID", filepath)
	if err.Error() != expectedMessage {
		t.Errorf("Expected error message '%s', but got '%s'", expectedMessage, err.Error())
	}
}

func TestFile_WithValidSystemTransactionFile(t *testing.T) {
	filepath := "../csv/system_transactions.csv"
	err := ValidateFile(filepath, "systemTransaction")
	if err != nil {
		t.Errorf("Expected no error for valid system transaction file, but got: %v", err)
	}
}

func TestFile_WithValidBankStatementFile(t *testing.T) {
	filepath := "../csv/bankA_20250605.csv"
	err := ValidateFile(filepath, "bankStatement")
	if err != nil {
		t.Errorf("Expected no error for valid bank statement file, but got: %v", err)
	}
}
