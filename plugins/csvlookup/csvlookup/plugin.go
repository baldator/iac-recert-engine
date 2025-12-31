package csvlookup

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/baldator/iac-recert-engine/pkg/api"
	"go.uber.org/zap"
)

// CSVLookupPlugin implements assignment based on CSV file lookup
type CSVLookupPlugin struct {
	csvFile    string
	keyRegex   *regexp.Regexp
	keyColumn  int
	valueColumn int
	logger     *zap.Logger
	csvData    map[string]string // key -> value mapping
}

// NewCSVLookupPlugin creates a new CSV lookup assignment plugin
func NewCSVLookupPlugin(logger *zap.Logger) api.AssignmentPlugin {
	return &CSVLookupPlugin{
		logger:  logger,
		csvData: make(map[string]string),
	}
}

func (p *CSVLookupPlugin) Init(config map[string]string) error {
	p.csvFile = config["csv_file"]
	if p.csvFile == "" {
		return fmt.Errorf("csv_file is required")
	}

	keyRegexStr := config["key_regex"]
	if keyRegexStr == "" {
		return fmt.Errorf("key_regex is required")
	}
	keyRegex, err := regexp.Compile(keyRegexStr)
	if err != nil {
		return fmt.Errorf("invalid key_regex pattern: %w", err)
	}
	p.keyRegex = keyRegex

	keyColumnStr := config["key_column"]
	if keyColumnStr == "" {
		return fmt.Errorf("key_column is required")
	}
	keyColumn, err := strconv.Atoi(keyColumnStr)
	if err != nil {
		return fmt.Errorf("invalid key_column: %w", err)
	}
	p.keyColumn = keyColumn

	valueColumnStr := config["value_column"]
	if valueColumnStr == "" {
		return fmt.Errorf("value_column is required")
	}
	valueColumn, err := strconv.Atoi(valueColumnStr)
	if err != nil {
		return fmt.Errorf("invalid value_column: %w", err)
	}
	p.valueColumn = valueColumn

	// Load CSV data
	if err := p.loadCSV(); err != nil {
		return fmt.Errorf("failed to load CSV file: %w", err)
	}

	return nil
}

func (p *CSVLookupPlugin) Resolve(files []api.FileInfo) (api.AssignmentResult, error) {
	// Extract key from files using regex
	key := p.extractKey(files)
	if key == "" {
		p.logger.Warn("no key found in files")
		return api.AssignmentResult{}, nil
	}

	// Look up value in CSV data
	value, exists := p.csvData[key]
	if !exists {
		p.logger.Warn("key not found in CSV", zap.String("key", key))
		return api.AssignmentResult{}, nil
	}

	return api.AssignmentResult{
		Assignees: []string{value},
	}, nil
}

func (p *CSVLookupPlugin) extractKey(files []api.FileInfo) string {
	for _, file := range files {
		content, err := p.readFile(file.Path)
		if err != nil {
			p.logger.Warn("failed to read file", zap.String("path", file.Path), zap.Error(err))
			continue
		}

		matches := p.keyRegex.FindStringSubmatch(content)
		if len(matches) > 1 {
			return matches[1]
		}
	}
	return ""
}

func (p *CSVLookupPlugin) readFile(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	var content strings.Builder
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		content.WriteString(scanner.Text() + "\n")
	}
	return content.String(), scanner.Err()
}

func (p *CSVLookupPlugin) loadCSV() error {
	file, err := os.Open(p.csvFile)
	if err != nil {
		return err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return err
	}

	if len(records) == 0 {
		return fmt.Errorf("CSV file is empty")
	}

	// Check if columns exist
	header := records[0]
	if p.keyColumn >= len(header) {
		return fmt.Errorf("key_column %d is out of range (header has %d columns)", p.keyColumn, len(header))
	}
	if p.valueColumn >= len(header) {
		return fmt.Errorf("value_column %d is out of range (header has %d columns)", p.valueColumn, len(header))
	}

	// Skip header if it exists (assume first row is header)
	startRow := 0
	if len(records) > 1 {
		startRow = 1
	}

	for i := startRow; i < len(records); i++ {
		record := records[i]
		if p.keyColumn >= len(record) || p.valueColumn >= len(record) {
			continue // Skip malformed rows
		}
		key := strings.TrimSpace(record[p.keyColumn])
		value := strings.TrimSpace(record[p.valueColumn])
		if key != "" && value != "" {
			p.csvData[key] = value
		}
	}

	if len(p.csvData) == 0 {
		return fmt.Errorf("no valid key-value pairs found in CSV")
	}

	p.logger.Info("loaded CSV data", zap.Int("entries", len(p.csvData)))
	return nil
}

// Export for plugin loading
var Plugin api.AssignmentPlugin = &CSVLookupPlugin{}
