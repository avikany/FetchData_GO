package ReadFiles

import (
	"bufio"
	"os"
	"strings"
)

func ReadFile(filePath string) ([][]string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines [][]string
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Split(line, ",")
		lines = append(lines, fields)

	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return lines, nil
}
