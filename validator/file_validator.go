package validator

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

var internalTransactionHeader = []string{"trxID", "amount", "type", "transactionTime"}
var bankStatementHeader = []string{"unique_identifier", "amount", "date"}

func ValidateFile(filePath string, fileType string) error {
	var file, err = os.OpenFile(filePath, os.O_RDONLY, 0644)
	if err != nil {
		return fmt.Errorf("open %s: no such file or directory", filePath)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	if condition := scanner.Scan(); !condition {
		return fmt.Errorf("file %s is empty", filePath)
	}

	header := strings.Split(scanner.Text(), ",")
	if err := validateHeader(header, fileType); err != nil {
		return fmt.Errorf("invalid header in %s: %v", filePath, err)
	}

	return nil
}

func validateHeader(header []string, fileType string) error {
	switch fileType {
	case "systemTransaction":
		if len(header) != len(internalTransactionHeader) {
			return fmt.Errorf("expected %d columns, got %d", len(internalTransactionHeader), len(header))
		}
		for i, col := range header {
			if col != internalTransactionHeader[i] {
				return fmt.Errorf("expected column %s, got %s", internalTransactionHeader[i], col)
			}
		}
		return nil
	case "bankStatement":
		if len(header) != len(bankStatementHeader) {
			return fmt.Errorf("expected %d columns, got %d", len(bankStatementHeader), len(header))
		}
		for i, col := range header {
			if col != bankStatementHeader[i] {
				return fmt.Errorf("expected column %s, got %s", bankStatementHeader[i], col)
			}
		}
		return nil
	default:
		return fmt.Errorf("unknown file type: %s", fileType)
	}

}
