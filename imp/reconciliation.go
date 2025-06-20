package impl

import (
	"math"
	"runtime"
	"sync"

	"github.com/sientong/reconciliation-service/model"
	"github.com/sientong/reconciliation-service/util"
)

func SimpleReconciliation() (*model.Output, error) {

	output := &model.Output{}

	// Check for a match between system transactions and bank statements
	for _, systemTransaction := range model.SystemTransactionRecords {
		for _, bankRecords := range model.BankStatementRecordsMap {
			for _, bankRecord := range bankRecords {

				if bankRecord.IsMatched || systemTransaction.Amount != math.Abs(bankRecord.Amount) {
					continue
				}

				bankRecordDate, err := util.ConvertBankStatementDate(bankRecord.Date)
				if err != nil {
					output.TotalInvalidRecords++
					continue
				}

				systemTransactionDate, err := util.ConvertSystemTransactionDate(systemTransaction.TransactionTime)
				if err != nil {
					output.TotalInvalidRecords++
					continue
				}

				if systemTransactionDate != bankRecordDate {
					continue
				}

				if systemTransaction.Type == "debit" && bankRecord.Amount > 0 {
					continue
				}

				if systemTransaction.Type == "credit" && bankRecord.Amount < 0 {
					continue
				}

				systemTransaction.IsMatched = true
				bankRecord.IsMatched = true
				output.TotalMatchedTransactions++
			}
		}

		// If no bank statement is matched with systm transaction, add to unmatched transaction
		if !systemTransaction.IsMatched {
			output.UnmatchedSystemTransactions = append(output.UnmatchedSystemTransactions, *systemTransaction)
			// fmt.Printf("Unmatched System Transaction: %s on %s\n", systemTransaction.TrxID, systemTransaction.TransactionTime)

			output.TotalDiscrepancies += math.Abs(systemTransaction.Amount)
			output.TotalUnmatchedSystemTransactions++
			output.TotalUnmatchedTransactions++
		}

		output.TotalProcessedRecords++
	}

	collectUnmatchedBankStmts(output)

	return output, nil
}

func ConcurrentReconcilliation() (*model.Output, error) {

	workers := 2 * runtime.NumCPU()

	// mutex to ensure no race condition for IsMatched field
	bankLocks := make(map[string]*sync.Mutex, len(model.BankStatementRecordsMap))
	for bankName := range model.BankStatementRecordsMap {
		bankLocks[bankName] = &sync.Mutex{}
	}

	// Initialize output structure
	output := &model.Output{}
	var outMu sync.Mutex

	// Create a channel to hold system transactions for processing
	jobs := make(chan *model.InternalTransactionRecord, len(model.SystemTransactionRecords))
	var wg sync.WaitGroup

	// Worker function to process transactions concurrently
	worker := func() {
		defer wg.Done()
		for systemTransaction := range jobs {
			processTransaction(systemTransaction, bankLocks, &outMu, output)
		}
	}

	// Start workers
	wg.Add(workers)
	for range workers {
		go worker()
	}

	// Send system transactions to the jobs channel
	// This will block until all transactions are sent
	for i := range model.SystemTransactionRecords {
		jobs <- model.SystemTransactionRecords[i]
	}

	close(jobs)
	wg.Wait()

	collectUnmatchedBankStmts(output)

	return output, nil
}

// processTransaction processes a single system transaction against bank records.
// It checks for matches, updates the records, and increments the output counters.
func processTransaction(transaction *model.InternalTransactionRecord,
	bankLocks map[string]*sync.Mutex,
	outMu *sync.Mutex,
	output *model.Output) {

	transactionDate, err := util.ConvertSystemTransactionDate(transaction.TransactionTime)
	if err != nil {
		outMu.Lock()
		output.TotalInvalidRecords++
		outMu.Unlock()
		return
	}

	// Check each bank's records for a match with the system transaction
	// Lock the bank records to prevent concurrent writes
	// and ensure thread safety
	for bankName, bankRecords := range model.BankStatementRecordsMap {
		lock := bankLocks[bankName]

		lock.Lock()
		for _, bankRecord := range bankRecords {
			if bankRecord.IsMatched || transaction.Amount != math.Abs(bankRecord.Amount) {
				continue
			}

			if transaction.Type == "debit" && bankRecord.Amount > 0 {
				continue
			}

			if transaction.Type == "credit" && bankRecord.Amount < 0 {
				continue
			}

			bankRecordDate, err := util.ConvertBankStatementDate(bankRecord.Date)
			if err != nil || transactionDate != bankRecordDate {
				outMu.Lock()
				output.TotalInvalidRecords++
				outMu.Unlock()
				continue
			}

			transaction.IsMatched = true
			bankRecord.IsMatched = true

			// Lock the output to update counters safely
			// Increment the matched transactions count
			outMu.Lock()
			output.TotalMatchedTransactions++
			outMu.Unlock()
			// fmt.Printf("Matched: System Transaction %s with Bank Record %s from %s status %t\n", transaction.TrxID, bankRecord.UniqueIdentifier, bankName, bankRecord.IsMatched)
			break
		}
		lock.Unlock()

		if transaction.IsMatched {
			break
		}
	}

	// Lock the output to update counters safely
	// Increment the total processed records count
	outMu.Lock()
	output.TotalProcessedRecords++
	outMu.Unlock()

	// If no bank statement is matched with system transaction, add to unmatched transaction
	if !transaction.IsMatched {
		output.UnmatchedSystemTransactions = append(output.UnmatchedSystemTransactions, *transaction)
		// fmt.Printf("Unmatched System Transaction: %s on %s\n", transaction.TrxID, transaction.TransactionTime)

		output.TotalDiscrepancies += math.Abs(transaction.Amount)
		output.TotalUnmatchedSystemTransactions++
		output.TotalUnmatchedTransactions++
	}
}

// Unprocessed bank statement is treated as unmatched
func collectUnmatchedBankStmts(output *model.Output) {
	for bankName, bankRecords := range model.BankStatementRecordsMap {
		for _, bankRecord := range bankRecords {
			if bankRecord.IsMatched {
				continue
			}

			if output.UnmatchedBankStmts == nil {
				output.UnmatchedBankStmts = make(map[string][]model.BankStatementRecord)
			}

			if _, exists := output.UnmatchedBankStmts[bankName]; !exists {
				output.UnmatchedBankStmts[bankName] = []model.BankStatementRecord{}
			}

			output.UnmatchedBankStmts[bankName] = append(output.UnmatchedBankStmts[bankName], *bankRecord)
			output.TotalDiscrepancies += math.Abs(bankRecord.Amount)
			output.TotalUnmatchedTransactions++
			output.TotalUnmatchedBankStmts++
			output.TotalProcessedRecords++
		}
	}
}
