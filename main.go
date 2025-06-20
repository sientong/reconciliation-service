package main

import (
	"fmt"
	"os"
	impl "reconciliation-service/imp"
	"reconciliation-service/model"
	"reconciliation-service/validator"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

func init() {
	// Initialize the model and other necessary components
	impl.InitModel()
	err := godotenv.Load()
	if err != nil {
		fmt.Println("Error loading .env file:", err)
	}
}

func main() {
	var argsRaw = os.Args[1:]
	if err := validator.ValidateArgs(argsRaw); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}

	fmt.Println("All arguments are valid. Proceeding with creating records...")

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

	fmt.Println("\nAll records created successfully. Starting reconciliation...")
	start := time.Now()

	reconcilliationStrategy := os.Getenv("RECONCILLIATION_STRATEGY")

	var output *model.Output
	var err error

	switch reconcilliationStrategy {
	case "simple":
		fmt.Println("Using simple reconciliation strategy...")
		output, err = impl.SimpleReconciliation()

	case "concurrent":
		fmt.Println("Using concurrent reconciliation strategy...")
		output, err = impl.ConcurrentReconcilliation()

	default:
		fmt.Printf("Unknown reconciliation strategy '%s', defaulting to 'concurrent'\n", reconcilliationStrategy)
		output, err = impl.ConcurrentReconcilliation()
	}

	if err != nil {
		fmt.Println("Error upon reconciliation:", err)
	}

	fmt.Println("\n---- Reconciliation results: ----")
	if output == nil {
		fmt.Println("No records to display.")
	} else {
		fmt.Printf("Total processed records: %d\n", output.TotalProcessedRecords)
		fmt.Printf("Total matched transactions: %d\n", output.TotalMatchedTransactions)
		fmt.Printf("Total unmatched transactions: %d\n", output.TotalUnmatchedTransactions)
		fmt.Printf("Total invalid records: %d\n", output.TotalInvalidRecords)
		fmt.Printf("Total discrepancies: %.2f\n\n", output.TotalDiscrepancies)
		fmt.Printf("Unmatched system transactions: %d\n", output.TotalUnmatchedSystemTransactions)
		for _, trx := range output.UnmatchedSystemTransactions {
			fmt.Printf(" - %s: %.2f on %s\n", trx.TrxID, trx.Amount, trx.TransactionTime)
		}
		fmt.Printf("Unmatched bank statements: %d\n", output.TotalUnmatchedBankStmts)
		for bankName, stmts := range output.UnmatchedBankStmts {
			fmt.Printf(" + %s\n", bankName)
			for _, stmt := range stmts {
				fmt.Printf("   - %s: %.2f on %s\n", stmt.UniqueIdentifier, stmt.Amount, stmt.Date)
			}
		}
	}

	duration := time.Since(start)
	fmt.Printf("\nReconciliation completed in %s\n", duration)
}
