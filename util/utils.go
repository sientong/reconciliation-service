package util

import "time"

func ConvertSystemTransactionDate(date string) (string, error) {
	// Assuming the date is in the format "2006-01-02T15:04:05Z" and we want to convert it to "20060102"
	parsedDate, err := time.Parse("2006-01-02T15:04:05Z", date)
	if err != nil {
		return "", err
	}
	return parsedDate.Format("20060102"), nil
}

func ConvertBankStatementDate(date string) (string, error) {
	// Assuming the date is in the format "2006-01-02" and we want to convert it to "20060102"
	parsedDate, err := time.Parse("2006-01-02", date)
	if err != nil {
		return "", err
	}
	return parsedDate.Format("20060102"), nil
}
