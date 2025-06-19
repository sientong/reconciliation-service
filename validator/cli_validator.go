package validator

import (
	"fmt"
	"strings"
	"time"
)

func ValidateArgs(args []string) error {

	// Check if the number of arguments is sufficient
	if len(args) < 4 {
		return fmt.Errorf("insufficient arguments provided: Expected at least 4 arguments, got %d", len(args))
	}

	// Validate file extension
	transactionFile := args[0]

	if len(transactionFile) < 4 || transactionFile[len(transactionFile)-4:] != ".csv" {
		return fmt.Errorf("invalid file format for system transactions: expected .csv, got %s", transactionFile)
	}

	bankStatementFiles := strings.Split(args[1], ",")
	for _, bankFile := range bankStatementFiles {
		if len(bankFile) < 4 || bankFile[len(bankFile)-4:] != ".csv" {
			return fmt.Errorf("invalid file format for bank statements: expected .csv, got %s", bankFile)
		}
	}

	layout := "20060102"

	// Validate the start date format
	startDate, err := time.Parse(layout, args[2])

	if err != nil {
		return fmt.Errorf("invalid start date format: %v. expected YYYYMMDD", err)
	}

	// Validate the end date format
	endDate, err := time.Parse(layout, args[3])

	if err != nil {
		return fmt.Errorf("invalid end date format: %v. expected YYYYMMDD", err)
	}

	// Validate that the end date is not earlier than the start date
	if startDate.After(endDate) {
		return fmt.Errorf("end date cannot be earlier than start date")
	}

	return nil
}
