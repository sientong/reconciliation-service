package main

import (
	"fmt"
	"os"
	impl "reconsiliation-service/imp"
	"reconsiliation-service/validator"
	"strings"
	"time"
)

func main() {
	var argsRaw = os.Args[1:]
	if err := validator.ValidateArgs(argsRaw); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}

	fmt.Println("All arguments are valid. Proceeding with the reconciliation process...")

	systemTransactionFile := argsRaw[0]
	bankStatementFiles := strings.Split(argsRaw[1], ",")
	startDate := argsRaw[2]
	endDate := argsRaw[3]

	if err := validator.ValidateFile(systemTransactionFile, "systemTransaction"); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}

	if err := impl.CreateRecords(systemTransactionFile, "systemTransaction", startDate, endDate); err != nil {
		fmt.Println("Error upon creating transaction records:", err)
	}

	for _, bankFile := range bankStatementFiles {
		if err := validator.ValidateFile(bankFile, "bankStatement"); err != nil {
			fmt.Fprintln(os.Stderr, "Error:", err)
			os.Exit(1)
		}

		if err := impl.CreateRecords(bankFile, "bankStatement", startDate, endDate); err != nil {
			fmt.Println("Error upon creating bank statement records:", err)
		}
	}

	fmt.Println("All records created successfully. Starting reconciliation...")
	start := time.Now()

	impl.Reconcile()

	duration := time.Since(start)
	fmt.Printf("Reconciliation process completed in %s\n", duration)
}
