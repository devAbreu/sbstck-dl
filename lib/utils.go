package lib

import (
	"bufio"
	"os"
)

func ReadLogFile(logFile string) (map[string]struct{}, error) {
	downloadedPosts := make(map[string]struct{})

	file, err := os.Open(logFile)
	if err != nil {
		if os.IsNotExist(err) {
			return downloadedPosts, nil
		}
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		downloadedPosts[scanner.Text()] = struct{}{}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return downloadedPosts, nil
}

func WriteLogFile(logFile string, downloadedPosts []string) error {
	file, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	for _, post := range downloadedPosts {
		if _, err := writer.WriteString(post + "\n"); err != nil {
			return err
		}
	}
	return writer.Flush()
}

// DateFilterFunc defines a function type for filtering dates.
type DateFilterFunc func(string) bool

// MakeDateFilterFunc creates a date filter function based on before and after dates.
func MakeDateFilterFunc(beforeDate string, afterDate string) DateFilterFunc {
	var dateFilterFunc DateFilterFunc
	if beforeDate != "" && afterDate != "" {
		dateFilterFunc = func(date string) bool {
			return date > afterDate && date < beforeDate
		}
	} else if beforeDate != "" {
		dateFilterFunc = func(date string) bool {
			return date < beforeDate
		}
	} else if afterDate != "" {
		dateFilterFunc = func(date string) bool {
			return date > afterDate
		}
	}
	return dateFilterFunc
}
