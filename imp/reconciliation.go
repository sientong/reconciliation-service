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

	// Bank locks to protect each bank’s records
	bankLocks := make(map[string]*sync.Mutex, len(model.BankStatementRecordsMap))
	for bankName := range model.BankStatementRecordsMap {
		bankLocks[bankName] = &sync.Mutex{}
	}

	jobs := make(chan *model.InternalTransactionRecord)
	results := make(chan *model.Output) // Per-worker results

	var wg sync.WaitGroup

	// Start workers
	wg.Add(workers)
	for range workers {
		go func() {
			defer wg.Done()
			localOutput := &model.Output{}

			for trx := range jobs {
				processTransactionLocal(trx, bankLocks, localOutput)
			}

			results <- localOutput // Send local results
		}()
	}

	// Feed jobs
	go func() {
		for _, trx := range model.SystemTransactionRecords {
			jobs <- trx
		}
		close(jobs)
	}()

	// Wait for workers to finish and close results channel
	go func() {
		wg.Wait()
		close(results)
	}()

	// Merge all local outputs into final output
	finalOutput := &model.Output{}

	for localOut := range results {
		mergeOutput(finalOutput, localOut)
	}

	// Collect unmatched bank statements (still single-threaded)
	collectUnmatchedBankStmts(finalOutput)

	return finalOutput, nil
}

// processTransaction processes a single system transaction against bank records.
// It checks for matches, updates the records, and increments the output counters.
func processTransactionLocal(
	transaction *model.InternalTransactionRecord,
	bankLocks map[string]*sync.Mutex,
	localOutput *model.Output) {

	transactionDate, err := util.ConvertSystemTransactionDate(transaction.TransactionTime)
	if err != nil {
		localOutput.TotalInvalidRecords++
		return
	}

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
				localOutput.TotalInvalidRecords++
				continue
			}

			transaction.IsMatched = true
			bankRecord.IsMatched = true

			localOutput.TotalMatchedTransactions++
			break
		}
		lock.Unlock()

		if transaction.IsMatched {
			break
		}
	}

	localOutput.TotalProcessedRecords++

	if !transaction.IsMatched {
		localOutput.UnmatchedSystemTransactions = append(localOutput.UnmatchedSystemTransactions, *transaction)
		localOutput.TotalDiscrepancies += math.Abs(transaction.Amount)
		localOutput.TotalUnmatchedSystemTransactions++
		localOutput.TotalUnmatchedTransactions++
	}
}

func mergeOutput(final *model.Output, local *model.Output) {
	final.TotalMatchedTransactions += local.TotalMatchedTransactions
	final.TotalProcessedRecords += local.TotalProcessedRecords
	final.TotalInvalidRecords += local.TotalInvalidRecords
	final.TotalUnmatchedTransactions += local.TotalUnmatchedTransactions
	final.TotalUnmatchedSystemTransactions += local.TotalUnmatchedSystemTransactions
	final.TotalDiscrepancies += local.TotalDiscrepancies

	// Combine unmatched slices
	final.UnmatchedSystemTransactions = append(final.UnmatchedSystemTransactions, local.UnmatchedSystemTransactions...)
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

func ConcurrentReconciliationIndexed() (*model.Output, error) {
	workers := 2 * runtime.NumCPU()
	index := BuildBankIndex()

	jobs := make(chan *model.InternalTransactionRecord)
	results := make(chan *model.Output)

	var wg sync.WaitGroup

	worker := func() {
		defer wg.Done()
		localOut := &model.Output{}

		for trx := range jobs {
			processTransactionIndexed(trx, &index, localOut)
		}

		results <- localOut
	}

	wg.Add(workers)
	for range workers {
		go worker()
	}

	// Feed jobs
	go func() {
		for _, trx := range model.SystemTransactionRecords {
			jobs <- trx
		}
		close(jobs)
	}()

	// Wait for workers, close results channel
	go func() {
		wg.Wait()
		close(results)
	}()

	finalOutput := &model.Output{}
	for res := range results {
		mergeOutput(finalOutput, res)
	}

	collectUnmatchedBankStmts(finalOutput)
	return finalOutput, nil
}

func processTransactionIndexed(
	transaction *model.InternalTransactionRecord,
	idx *MatchIndex,
	localOutput *model.Output) {

	transactionDate, err := util.ConvertSystemTransactionDate(transaction.TransactionTime)
	if err != nil {
		localOutput.TotalInvalidRecords++
		return
	}

	amount := math.Abs(transaction.Amount)
	txType := transaction.Type

	// Lookup possible matches
	dateBucket, ok := idx.Index[transactionDate]
	if !ok {
		localOutput.TotalUnmatchedSystemTransactions++
		localOutput.TotalProcessedRecords++
		localOutput.UnmatchedSystemTransactions = append(localOutput.UnmatchedSystemTransactions, *transaction)
		localOutput.TotalDiscrepancies += amount
		localOutput.TotalUnmatchedTransactions++
		return
	}

	amtBucket, ok := dateBucket[amount]
	if !ok {
		localOutput.TotalUnmatchedSystemTransactions++
		localOutput.TotalProcessedRecords++
		localOutput.UnmatchedSystemTransactions = append(localOutput.UnmatchedSystemTransactions, *transaction)
		localOutput.TotalDiscrepancies += amount
		localOutput.TotalUnmatchedTransactions++
		return
	}

	typeBucket, ok := amtBucket[txType]
	if !ok {
		localOutput.TotalUnmatchedSystemTransactions++
		localOutput.TotalProcessedRecords++
		localOutput.UnmatchedSystemTransactions = append(localOutput.UnmatchedSystemTransactions, *transaction)
		localOutput.TotalDiscrepancies += amount
		localOutput.TotalUnmatchedTransactions++
		return
	}

	// Lock the bucket for safe matching
	bucketLock := idx.Locks[transactionDate][amount][txType]
	bucketLock.Lock()
	defer bucketLock.Unlock()

	for _, rec := range typeBucket {
		if !rec.IsMatched {
			rec.IsMatched = true
			transaction.IsMatched = true
			localOutput.TotalMatchedTransactions++
			break
		}
	}

	localOutput.TotalProcessedRecords++
	if !transaction.IsMatched {
		localOutput.TotalUnmatchedSystemTransactions++
		localOutput.TotalUnmatchedTransactions++
		localOutput.TotalDiscrepancies += amount
		localOutput.UnmatchedSystemTransactions = append(localOutput.UnmatchedSystemTransactions, *transaction)
	}
}
