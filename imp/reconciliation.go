package impl

import (
	"math"
	"reconsiliation-service/model"
	"reconsiliation-service/util"
	"runtime"
	"sync"
)

func SimpleReconciliation() (*model.Output, error) {

	output := &model.Output{}

	// Check for a match between system transactions and bank statements
	for _, systemTransaction := range model.SystemTransactionRecords {
		for _, bankRecords := range model.BankStatementRecordsMap {
			for i, bankRecord := range bankRecords {

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

				if !bankRecord.IsMatched && systemTransaction.Amount == math.Abs(bankRecord.Amount) && systemTransactionDate == bankRecordDate {
					systemTransaction.IsMatched = true
					bankRecords[i].IsMatched = true
					// fmt.Printf("Matched: System Transaction %s with Bank Record %s from %s\n", systemTransaction.TrxID, bankRecord.UniqueIdentifier, bankName)
					output.TotalMatchedTransactions++
					break
				}
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

	output := &model.Output{}
	var outMu sync.Mutex

	jobs := make(chan *model.InternalTransactionRecord, len(model.SystemTransactionRecords))
	var wg sync.WaitGroup

	worker := func() {
		defer wg.Done()
		for systemTransaction := range jobs {
			processTransaction(systemTransaction, bankLocks, &outMu, output)
		}
	}

	wg.Add(workers)
	for i := 0; i < workers; i++ {
		go worker()
	}

	for i := range model.SystemTransactionRecords {
		jobs <- model.SystemTransactionRecords[i]
	}

	close(jobs)
	wg.Wait()

	collectUnmatchedBankStmts(output)

	return output, nil
}

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

	for bankName, bankRecords := range model.BankStatementRecordsMap {
		lock := bankLocks[bankName]

		lock.Lock()
		for _, bankRecord := range bankRecords {
			if bankRecord.IsMatched || transaction.Amount != math.Abs(bankRecord.Amount) {
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

	outMu.Lock()
	defer outMu.Unlock()
	output.TotalProcessedRecords++
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
			if !bankRecord.IsMatched {
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
}
