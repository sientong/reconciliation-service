package test

import (
	. "reconsiliation-service/validator"
	"testing"
)

func TestCLI_WithValidArguments(t *testing.T) {
	args := []string{"system_transactions.csv", "bank_statements.csv", "20230101", "20231231"}

	err := ValidateArgs(args)
	if err != nil {
		t.Errorf("Expected no error for valid arguments, but got: %v", err)
	}
}

func TestCLI_WithInvalidStartDate(t *testing.T) {
	args := []string{"system_transactions.csv", "bank_statements.csv", "2023-01-01", "20231231"}

	err := ValidateArgs(args)
	if err == nil {
		t.Errorf("Expected an error for invalid arguments, but got nil")
	}

	expectedMessage := "invalid start date format: parsing time \"2023-01-01\" as \"20060102\": cannot parse \"-01-01\" as \"01\". expected YYYYMMDD"
	if err.Error() != expectedMessage {
		t.Errorf("Expected error message '%s', but got '%s'", expectedMessage, err.Error())
	}
}

func TestCLI_WithInvalidEndDate(t *testing.T) {
	args := []string{"system_transactions.csv", "bank_statements.csv", "20230101", "2023-12-31"}

	err := ValidateArgs(args)
	if err == nil {
		t.Errorf("Expected an error for invalid arguments, but got nil")
	}

	expectedMessage := "invalid end date format: parsing time \"2023-12-31\" as \"20060102\": cannot parse \"-12-31\" as \"01\". expected YYYYMMDD"
	if err.Error() != expectedMessage {
		t.Errorf("Expected error message '%s', but got '%s'", expectedMessage, err.Error())
	}
}

func TestCLI_WithEndDateGreaterThanStartDate(t *testing.T) {
	args := []string{"system_transactions.csv", "bank_statements.csv", "20230201", "20230101"}

	err := ValidateArgs(args)
	if err == nil {
		t.Errorf("Expected an error for end date greater than start date, but got nil")
	}

	expectedMessage := "end date cannot be earlier than start date"
	if err.Error() != expectedMessage {
		t.Errorf("Expected error message '%s', but got '%s'", expectedMessage, err.Error())
	}
}

func TestCLI_WithMissingArguments(t *testing.T) {
	args := []string{"system_transactions.csv", "bank_statements.csv"}
	err := ValidateArgs(args)

	if err == nil {
		t.Errorf("Expected an error for missing arguments, but got nil")
	}

	expectedMessage := "insufficient arguments provided: Expected at least 4 arguments, got 2"
	if err.Error() != expectedMessage {
		t.Errorf("Expected error message '%s', but got '%s'", expectedMessage, err.Error())
	}
}

func TestCLI_WithInvalidTransactionFileExtension(t *testing.T) {
	args := []string{"system_transactions", "bank_statements.csv", "20230101", "20231231"}

	err := ValidateArgs(args)
	if err == nil {
		t.Errorf("Expected error for invalid arguments, but got: %v", err)
	}
	expectedMessage := "invalid file format for system transactions: expected .csv, got system_transactions"
	if err.Error() != expectedMessage {
		t.Errorf("Expected error message '%s', but got '%s'", expectedMessage, err.Error())
	}
}

func TestCLI_WithInvalidBankStatementFileExtension(t *testing.T) {
	args := []string{"system_transactions.csv", "bank_statements.csv,bank_statements", "20230101", "20231231"}

	err := ValidateArgs(args)
	if err == nil {
		t.Errorf("Expected error for invalid arguments, but got: %v", err)
	}
	expectedMessage := "invalid file format for bank statements: expected .csv, got bank_statements"
	if err.Error() != expectedMessage {
		t.Errorf("Expected error message '%s', but got '%s'", expectedMessage, err.Error())
	}
}
