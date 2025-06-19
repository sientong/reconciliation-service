package impl

import (
	"encoding/csv"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"reconsiliation-service/model"
	"reconsiliation-service/util"
	validator "reconsiliation-service/validator"
	"strconv"
	"strings"
)

func CreateRecords(filePath string, recordType string, startDate string, endDate string) error {

	switch recordType {
	case "systemTransaction":
		if err := createSystemTransactionsRecords(filePath, startDate, endDate); err != nil {
			return fmt.Errorf("failed to create system transactions records: %w", err)
		}
	case "bankStatement":
		if err := createBankStatementRecords(filePath, startDate, endDate); err != nil {
			return fmt.Errorf("failed to create bank statement records: %w", err)
		}
	default:
		return fmt.Errorf("unknown record type: %s", recordType)
	}

	return nil
}

func createSystemTransactionsRecords(filePath string, startDate string, endDate string) error {
	fmt.Println("Creating system transaction records from:", filePath)

	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("open %s: no such file or directory", filePath)
	}
	defer file.Close()

	csvReader := csv.NewReader(file)
	csvRecords, err := csvReader.ReadAll()
	if err != nil {
		return fmt.Errorf("read %s: %w", filePath, err)
	}

	for _, row := range csvRecords[1:] { // Skip header row
		err := parseSystemTransactionRecord(row, startDate, endDate)
		if err != nil {
			fmt.Printf("error parsing record %v: %v\n", row, err)
			continue
		}
	}

	return nil
}

func parseSystemTransactionRecord(record []string, startDate string, endDate string) error {
	err := validator.ValidateRecord(record, "systemTransaction")
	if err != nil {
		return fmt.Errorf("validate record %v: %w", record, err)
	}

	amount, err := strconv.ParseFloat(record[1], 64)
	if err != nil {
		return fmt.Errorf("parse amount %s: %w", record[1], err)
	}

	newRecord := &model.InternalTransactionRecord{
		TrxID:           record[0],
		Amount:          amount,
		Type:            record[2],
		TransactionTime: record[3],
		IsMatched:       false,
	}

	transactionDate, err := util.ConvertSystemTransactionDate(newRecord.TransactionTime)
	if err != nil {
		return fmt.Errorf("error when converting transaction date %s: %w", newRecord.TransactionTime, err)
	}

	if transactionDate < startDate || transactionDate > endDate {
		return fmt.Errorf("transaction time %s is out of range [%s, %s]", newRecord.TransactionTime, startDate, endDate)
	}

	model.SystemTransactionRecords = append(model.SystemTransactionRecords, *newRecord)

	return nil
}

func createBankStatementRecords(filePath string, startDate string, endDate string) error {
	fmt.Println("Creating bank statement records from:", filePath)

	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("open %s: no such file or directory", filePath)
	}
	defer file.Close()

	bankName := strings.Split(filepath.Base(filePath), "_")[0] // Get bank name from file name

	csvReader := csv.NewReader(file)
	csvRecords, err := csvReader.ReadAll()
	if err != nil {
		return fmt.Errorf("read %s: %w", filePath, err)
	}

	for _, row := range csvRecords[1:] { // Skip header row
		err := parseBankStatementRecord(row, bankName, startDate, endDate)
		if err != nil {
			fmt.Printf("error parsing record %v: %v\n", row, err)
			continue
		}
	}

	return nil
}

func parseBankStatementRecord(record []string, bankName string, startDate string, endDate string) error {
	err := validator.ValidateRecord(record, "bankStatement")
	if err != nil {
		return fmt.Errorf("validate record %v: %w", record, err)
	}

	amount, err := strconv.ParseFloat(record[1], 64)
	if err != nil {
		return fmt.Errorf("parse amount %s: %w", record[1], err)
	}

	newRecord := &model.BankStatementRecord{
		UniqueIdentifier: record[0],
		Amount:           amount,
		Date:             record[2],
		IsMatched:        false,
	}

	transactionDate, err := util.ConvertBankStatementDate(newRecord.Date)
	if err != nil {
		return fmt.Errorf("error when converting bank statement date %s: %w", newRecord.Date, err)
	}

	if transactionDate < startDate || transactionDate > endDate {
		return fmt.Errorf("date %s is out of range [%s, %s]", newRecord.Date, startDate, endDate)
	}

	if _, exists := model.BankStatementRecordsMap[bankName]; !exists {
		model.BankStatementRecordsMap[bankName] = []model.BankStatementRecord{}
	}

	model.BankStatementRecordsMap[bankName] = append(model.BankStatementRecordsMap[bankName], *newRecord)

	return nil
}

func Reconcile() (*model.Output, error) {

	output := &model.Output{}

	// Check for a match between system transactions and bank statements
	for _, systemTransaction := range model.SystemTransactionRecords {
		for bankName, bankRecords := range model.BankStatementRecordsMap {
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
					fmt.Printf("Matched: System Transaction %s with Bank Record %s from %s\n", systemTransaction.TrxID, bankRecord.UniqueIdentifier, bankName)
					output.TotalMatchedTransactions++
					break
				}
			}
		}

		// If no bank statement is matched with systm transaction, add to unmatched transaction
		if !systemTransaction.IsMatched {
			output.UnmatchedSystemTransactions = append(output.UnmatchedSystemTransactions, systemTransaction)
			fmt.Printf("Unmatched System Transaction: %s on %s\n", systemTransaction.TrxID, systemTransaction.TransactionTime)

			output.TotalDiscrepancies += math.Abs(systemTransaction.Amount)
			output.TotalUnmatchedTransactions++
		}

		output.TotalProcessedRecords++
	}

	// Unprocessed bank statement is treated as unmatched
	for bankName, bankRecords := range model.BankStatementRecordsMap {
		for _, bankRecord := range bankRecords {
			if !bankRecord.IsMatched {
				if output.UnmatchedBankStmts == nil {
					output.UnmatchedBankStmts = make(map[string][]model.BankStatementRecord)
				}

				if _, exists := output.UnmatchedBankStmts[bankName]; !exists {
					output.UnmatchedBankStmts[bankName] = []model.BankStatementRecord{}
				}

				output.UnmatchedBankStmts[bankName] = append(output.UnmatchedBankStmts[bankName], bankRecord)
				fmt.Printf("Unmatched Bank Statement: %s on %.2f from %s on %s\n", bankRecord.UniqueIdentifier, bankRecord.Amount, bankName, bankRecord.Date)

				output.TotalDiscrepancies += math.Abs(bankRecord.Amount)
				output.TotalUnmatchedTransactions++
				output.TotalProcessedRecords++
			}
		}
	}
	return output, nil
}
