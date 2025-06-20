package impl

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"reconsiliation-service/model"
	"reconsiliation-service/util"
	validator "reconsiliation-service/validator"
	"strconv"
	"strings"
)

func InitModel() {
	// Initialize the model and other necessary components
	model.SystemTransactionRecords = []*model.InternalTransactionRecord{}
	model.BankStatementRecordsMap = make(map[string][]*model.BankStatementRecord)
}

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

	model.SystemTransactionRecords = append(model.SystemTransactionRecords, newRecord)

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
		model.BankStatementRecordsMap[bankName] = []*model.BankStatementRecord{}
	}

	model.BankStatementRecordsMap[bankName] = append(model.BankStatementRecordsMap[bankName], newRecord)

	return nil
}
