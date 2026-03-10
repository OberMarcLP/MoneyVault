package services

import (
	"bufio"
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"moneyvault/internal/encryption"
	"moneyvault/internal/models"
	"moneyvault/internal/repositories"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type ImportService struct {
	importRepo      *repositories.ImportRepository
	transactionRepo *repositories.TransactionRepository
	categoryRepo    *repositories.CategoryRepository
	enc             *encryption.Service
}

func NewImportService(
	importRepo *repositories.ImportRepository,
	transactionRepo *repositories.TransactionRepository,
	categoryRepo *repositories.CategoryRepository,
	enc *encryption.Service,
) *ImportService {
	return &ImportService{
		importRepo:      importRepo,
		transactionRepo: transactionRepo,
		categoryRepo:    categoryRepo,
		enc:             enc,
	}
}

func stripBOM(data []byte) []byte {
	if len(data) >= 3 && data[0] == 0xEF && data[1] == 0xBB && data[2] == 0xBF {
		return data[3:]
	}
	return data
}

func detectDelimiter(data []byte) rune {
	scanner := bufio.NewScanner(bytes.NewReader(data))
	if !scanner.Scan() {
		return ','
	}
	firstLine := scanner.Text()

	for _, delim := range []rune{',', ';', '\t', '|'} {
		if strings.ContainsRune(firstLine, delim) {
			parts := strings.Split(firstLine, string(delim))
			if len(parts) >= 2 {
				return delim
			}
		}
	}
	return ','
}

func newCSVReader(data []byte) *csv.Reader {
	cleaned := stripBOM(data)
	delim := detectDelimiter(cleaned)

	reader := csv.NewReader(bytes.NewReader(cleaned))
	reader.Comma = delim
	reader.LazyQuotes = true
	reader.TrimLeadingSpace = true
	reader.FieldsPerRecord = -1
	return reader
}

func (s *ImportService) PreviewCSV(data io.Reader) (*models.CSVPreview, error) {
	raw, err := io.ReadAll(data)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}
	if len(raw) == 0 {
		return nil, fmt.Errorf("file is empty")
	}
	if !utf8.Valid(raw) {
		return nil, fmt.Errorf("file encoding not supported; please save as UTF-8")
	}

	reader := newCSVReader(raw)

	headers, err := reader.Read()
	if err != nil {
		return nil, fmt.Errorf("failed to read CSV headers: %w", err)
	}
	if len(headers) < 2 {
		delim := detectDelimiter(stripBOM(raw))
		return nil, fmt.Errorf("only %d column(s) detected (delimiter=%q); expected at least 2 — check that your CSV uses the right separator", len(headers), string(delim))
	}

	for i := range headers {
		headers[i] = strings.TrimSpace(headers[i])
	}

	var rows []models.CSVPreviewRow
	total := 0
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			continue
		}
		total++
		if len(rows) < 10 {
			values := make(map[string]string)
			for i, h := range headers {
				if i < len(record) {
					values[h] = record[i]
				}
			}
			rows = append(rows, models.CSVPreviewRow{Values: values})
		}
	}

	if total == 0 {
		return nil, fmt.Errorf("CSV has headers but no data rows")
	}

	return &models.CSVPreview{
		Headers: headers,
		Rows:    rows,
		Total:   total,
	}, nil
}

func (s *ImportService) ImportCSV(
	userID uuid.UUID,
	accountID uuid.UUID,
	data io.Reader,
	mapping models.ColumnMapping,
	filename string,
	postedOnly bool,
) (*models.ImportJob, error) {
	if _, err := s.enc.GetDEK(userID); err != nil {
		return nil, fmt.Errorf("encryption key not available — please log out and log back in")
	}

	raw, err := io.ReadAll(data)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	reader := newCSVReader(raw)

	headers, err := reader.Read()
	if err != nil {
		return nil, fmt.Errorf("failed to read CSV headers: %w", err)
	}

	for i := range headers {
		headers[i] = strings.TrimSpace(headers[i])
	}

	headerIdx := make(map[string]int)
	for i, h := range headers {
		headerIdx[h] = i
	}

	dateIdx, ok := headerIdx[mapping.Date]
	if !ok {
		return nil, fmt.Errorf("date column %q not found in headers %v", mapping.Date, headers)
	}
	amountIdx, ok := headerIdx[mapping.Amount]
	if !ok {
		return nil, fmt.Errorf("amount column %q not found in headers %v", mapping.Amount, headers)
	}

	descIdx := -1
	if mapping.Description != "" {
		if idx, ok := headerIdx[mapping.Description]; ok {
			descIdx = idx
		}
	}
	merchantIdx := -1
	if mapping.Merchant != "" {
		if idx, ok := headerIdx[mapping.Merchant]; ok {
			merchantIdx = idx
		}
	}
	categoryIdx := -1
	if mapping.Category != "" {
		if idx, ok := headerIdx[mapping.Category]; ok {
			categoryIdx = idx
		}
	}
	subCategoryIdx := -1
	if mapping.SubCategory != "" {
		if idx, ok := headerIdx[mapping.SubCategory]; ok {
			subCategoryIdx = idx
		}
	}
	typeIdx := -1
	if mapping.Type != "" {
		if idx, ok := headerIdx[mapping.Type]; ok {
			typeIdx = idx
		}
	}
	statusIdx := -1
	if mapping.Status != "" {
		if idx, ok := headerIdx[mapping.Status]; ok {
			statusIdx = idx
		}
	}
	currencyIdx := -1
	if mapping.Currency != "" {
		if idx, ok := headerIdx[mapping.Currency]; ok {
			currencyIdx = idx
		}
	}

	job := &models.ImportJob{
		ID:            uuid.New(),
		UserID:        userID,
		AccountID:     accountID,
		Filename:      filename,
		Status:        models.ImportProcessing,
		ColumnMapping: map[string]string{
			"date":         mapping.Date,
			"amount":       mapping.Amount,
			"description":  mapping.Description,
			"merchant":     mapping.Merchant,
			"category":     mapping.Category,
			"sub_category": mapping.SubCategory,
			"type":         mapping.Type,
			"status":       mapping.Status,
			"currency":     mapping.Currency,
		},
	}

	var records [][]string
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			continue
		}
		records = append(records, record)
	}
	job.TotalRows = len(records)

	if job.TotalRows == 0 {
		job.Status = models.ImportCompleted
		_ = s.importRepo.Create(job)
		return job, nil
	}

	existingList, _, _ := s.transactionRepo.List(userID, models.TransactionFilter{AccountID: &accountID, PerPage: 10000})

	type decryptedTx struct {
		Date   time.Time
		Amount float64
		Desc   string
	}
	var existingDecrypted []decryptedTx
	for _, tx := range existingList {
		decAmt, err := s.enc.DecryptField(userID, tx.Amount)
		if err != nil {
			continue
		}
		amt, err := strconv.ParseFloat(decAmt, 64)
		if err != nil {
			continue
		}
		decDesc, _ := s.enc.DecryptField(userID, tx.Description)
		existingDecrypted = append(existingDecrypted, decryptedTx{Date: tx.Date, Amount: amt, Desc: decDesc})
	}

	categories, _ := s.categoryRepo.ListByUser(userID)
	categoryCache := make(map[string]uuid.UUID)
	for _, c := range categories {
		key := categoryCacheKey(c.Type, c.Name)
		categoryCache[key] = c.ID
	}

	imported, duplicates, skipped := 0, 0, 0

	// Wrap all inserts in a single database transaction for atomicity
	txErr := repositories.WithTransaction(s.transactionRepo.DB(), func(dbTx *sqlx.Tx) error {
		for _, record := range records {
			if dateIdx >= len(record) || amountIdx >= len(record) {
				skipped++
				continue
			}

			dateStr := strings.TrimSpace(record[dateIdx])
			amountStr := strings.TrimSpace(record[amountIdx])
			description := ""
			if descIdx >= 0 && descIdx < len(record) {
				description = sanitizeCSVValue(strings.TrimSpace(record[descIdx]))
			}
			merchant := ""
			if merchantIdx >= 0 && merchantIdx < len(record) {
				merchant = sanitizeCSVValue(strings.TrimSpace(record[merchantIdx]))
			}
			categoryName := ""
			if categoryIdx >= 0 && categoryIdx < len(record) {
				categoryName = sanitizeCSVValue(strings.TrimSpace(record[categoryIdx]))
			}
			subCategoryName := ""
			if subCategoryIdx >= 0 && subCategoryIdx < len(record) {
				subCategoryName = sanitizeCSVValue(strings.TrimSpace(record[subCategoryIdx]))
			}
			typeRaw := ""
			if typeIdx >= 0 && typeIdx < len(record) {
				typeRaw = strings.TrimSpace(record[typeIdx])
			}
			statusRaw := ""
			if statusIdx >= 0 && statusIdx < len(record) {
				statusRaw = strings.TrimSpace(record[statusIdx])
			}
			currencyRaw := ""
			if currencyIdx >= 0 && currencyIdx < len(record) {
				currencyRaw = strings.TrimSpace(record[currencyIdx])
			}

			if postedOnly && statusRaw != "" && !strings.EqualFold(statusRaw, "posted") {
				skipped++
				continue
			}

			parsedDate := parseDate(dateStr)
			if parsedDate.IsZero() {
				skipped++
				continue
			}

			parsedAmount, err := parseAmount(amountStr)
			if err != nil {
				skipped++
				continue
			}

			txType := inferTransactionType(typeRaw, parsedAmount)
			absAmount := math.Abs(parsedAmount)
			currency := normalizeCurrencyCode(currencyRaw)
			if currency == "" {
				currency = "USD"
			}

			descParts := make([]string, 0, 2)
			if merchant != "" {
				descParts = append(descParts, merchant)
			}
			if description != "" {
				descParts = append(descParts, description)
			}
			if len(descParts) == 0 {
				description = "Imported transaction"
			} else {
				description = strings.Join(descParts, " | ")
			}

			isDup := false
			for _, ex := range existingDecrypted {
				if ex.Date.Equal(parsedDate) && math.Abs(ex.Amount-absAmount) < 0.01 && ex.Desc == description {
					isDup = true
					break
				}
			}
			if isDup {
				duplicates++
				continue
			}

			encAmount, err := s.enc.EncryptField(userID, fmt.Sprintf("%.2f", absAmount))
			if err != nil {
				skipped++
				continue
			}
			encDesc, err := s.enc.EncryptField(userID, description)
			if err != nil {
				skipped++
				continue
			}

			tx := &models.Transaction{
				ID:           uuid.New(),
				AccountID:    accountID,
				UserID:       userID,
				Type:         txType,
				Amount:       encAmount,
				Currency:     currency,
				Description:  encDesc,
				Date:         parsedDate,
				Tags:         json.RawMessage(`[]`),
				ImportSource: "csv",
			}

			catCandidate := firstNonEmpty(categoryName, subCategoryName)
			if catCandidate != "" {
				catID, err := s.resolveOrCreateCategoryTx(dbTx, userID, txType, catCandidate, categoryCache)
				if err == nil {
					tx.CategoryID = &catID
				}
			}

			if err := s.transactionRepo.CreateWithTx(dbTx, tx); err != nil {
				log.Printf("CSV import: failed to create transaction row %d: %v", imported+duplicates+skipped+1, err)
				skipped++
				continue
			}
			imported++
		}

		// Create import job record within the same transaction
		job.ImportedRows = imported
		job.DuplicateRows = duplicates
		job.Status = models.ImportCompleted
		if imported == 0 && skipped == job.TotalRows {
			errMsg := fmt.Sprintf("all %d rows skipped — check date/amount format", skipped)
			job.ErrorMessage = &errMsg
			job.Status = models.ImportFailed
		}
		return s.importRepo.CreateWithTx(dbTx, job)
	})

	if txErr != nil {
		job.Status = models.ImportFailed
		errMsg := txErr.Error()
		job.ErrorMessage = &errMsg
		_ = s.importRepo.Create(job)
		return job, txErr
	}

	return job, nil
}

func (s *ImportService) ListJobs(userID uuid.UUID) ([]models.ImportJob, error) {
	return s.importRepo.List(userID)
}

// PreviewOFX parses an OFX file and returns a preview (reusing CSV preview struct).
func (s *ImportService) PreviewOFX(data io.Reader) (*models.CSVPreview, error) {
	rows, err := parseOFX(data)
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, fmt.Errorf("no transactions found in OFX file")
	}

	headers := []string{"Date", "Amount", "Description", "Type"}
	var previewRows []models.CSVPreviewRow
	for i, r := range rows {
		if i >= 10 {
			break
		}
		previewRows = append(previewRows, models.CSVPreviewRow{Values: map[string]string{
			"Date":        r.Date,
			"Amount":      r.Amount,
			"Description": r.Description,
			"Type":        r.Type,
		}})
	}

	return &models.CSVPreview{
		Headers: headers,
		Rows:    previewRows,
		Total:   len(rows),
	}, nil
}

// ImportOFX imports transactions from an OFX file.
func (s *ImportService) ImportOFX(userID, accountID uuid.UUID, data io.Reader, filename string) (*models.ImportJob, error) {
	if _, err := s.enc.GetDEK(userID); err != nil {
		return nil, fmt.Errorf("encryption key not available — please log out and log back in")
	}

	rows, err := parseOFX(data)
	if err != nil {
		return nil, err
	}

	return s.importParsedRows(userID, accountID, rows, filename, "ofx")
}

// PreviewQIF parses a QIF file and returns a preview.
func (s *ImportService) PreviewQIF(data io.Reader) (*models.CSVPreview, error) {
	rows, err := parseQIF(data)
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, fmt.Errorf("no transactions found in QIF file")
	}

	headers := []string{"Date", "Amount", "Description", "Type"}
	var previewRows []models.CSVPreviewRow
	for i, r := range rows {
		if i >= 10 {
			break
		}
		previewRows = append(previewRows, models.CSVPreviewRow{Values: map[string]string{
			"Date":        r.Date,
			"Amount":      r.Amount,
			"Description": r.Description,
			"Type":        r.Type,
		}})
	}

	return &models.CSVPreview{
		Headers: headers,
		Rows:    previewRows,
		Total:   len(rows),
	}, nil
}

// ImportQIF imports transactions from a QIF file.
func (s *ImportService) ImportQIF(userID, accountID uuid.UUID, data io.Reader, filename string) (*models.ImportJob, error) {
	if _, err := s.enc.GetDEK(userID); err != nil {
		return nil, fmt.Errorf("encryption key not available — please log out and log back in")
	}

	rows, err := parseQIF(data)
	if err != nil {
		return nil, err
	}

	return s.importParsedRows(userID, accountID, rows, filename, "qif")
}

type parsedRow struct {
	Date        string
	Amount      string
	Description string
	Type        string
}

func (s *ImportService) importParsedRows(userID, accountID uuid.UUID, rows []parsedRow, filename, source string) (*models.ImportJob, error) {
	job := &models.ImportJob{
		ID:        uuid.New(),
		UserID:    userID,
		AccountID: accountID,
		Filename:  filename,
		Status:    models.ImportProcessing,
		TotalRows: len(rows),
		ColumnMapping: map[string]string{
			"format": source,
		},
	}

	if len(rows) == 0 {
		job.Status = models.ImportCompleted
		_ = s.importRepo.Create(job)
		return job, nil
	}

	// Duplicate detection
	existingList, _, _ := s.transactionRepo.List(userID, models.TransactionFilter{AccountID: &accountID, PerPage: 10000})
	type decryptedTx struct {
		Date   time.Time
		Amount float64
		Desc   string
	}
	var existingDecrypted []decryptedTx
	for _, tx := range existingList {
		decAmt, err := s.enc.DecryptField(userID, tx.Amount)
		if err != nil {
			continue
		}
		amt, err := strconv.ParseFloat(decAmt, 64)
		if err != nil {
			continue
		}
		decDesc, _ := s.enc.DecryptField(userID, tx.Description)
		existingDecrypted = append(existingDecrypted, decryptedTx{Date: tx.Date, Amount: amt, Desc: decDesc})
	}

	imported, duplicates, skipped := 0, 0, 0

	txErr := repositories.WithTransaction(s.transactionRepo.DB(), func(dbTx *sqlx.Tx) error {
		for _, r := range rows {
			parsedDate := parseDate(r.Date)
			if parsedDate.IsZero() {
				skipped++
				continue
			}
			parsedAmount, err := parseAmount(r.Amount)
			if err != nil {
				skipped++
				continue
			}

			txType := inferTransactionType(r.Type, parsedAmount)
			absAmount := math.Abs(parsedAmount)
			description := r.Description
			if description == "" {
				description = "Imported transaction"
			}

			isDup := false
			for _, ex := range existingDecrypted {
				if ex.Date.Equal(parsedDate) && math.Abs(ex.Amount-absAmount) < 0.01 && ex.Desc == description {
					isDup = true
					break
				}
			}
			if isDup {
				duplicates++
				continue
			}

			encAmount, err := s.enc.EncryptField(userID, fmt.Sprintf("%.2f", absAmount))
			if err != nil {
				skipped++
				continue
			}
			encDesc, err := s.enc.EncryptField(userID, description)
			if err != nil {
				skipped++
				continue
			}

			tx := &models.Transaction{
				ID:           uuid.New(),
				AccountID:    accountID,
				UserID:       userID,
				Type:         txType,
				Amount:       encAmount,
				Currency:     "USD",
				Description:  encDesc,
				Date:         parsedDate,
				Tags:         json.RawMessage(`[]`),
				ImportSource: source,
			}

			if err := s.transactionRepo.CreateWithTx(dbTx, tx); err != nil {
				skipped++
				continue
			}
			imported++
		}

		job.ImportedRows = imported
		job.DuplicateRows = duplicates
		job.Status = models.ImportCompleted
		if imported == 0 && skipped == job.TotalRows {
			errMsg := fmt.Sprintf("all %d rows skipped — check file format", skipped)
			job.ErrorMessage = &errMsg
			job.Status = models.ImportFailed
		}
		return s.importRepo.CreateWithTx(dbTx, job)
	})

	if txErr != nil {
		job.Status = models.ImportFailed
		errMsg := txErr.Error()
		job.ErrorMessage = &errMsg
		_ = s.importRepo.Create(job)
		return job, txErr
	}

	return job, nil
}

// parseOFX parses OFX/QFX files (SGML-based financial data exchange format).
func parseOFX(data io.Reader) ([]parsedRow, error) {
	raw, err := io.ReadAll(data)
	if err != nil {
		return nil, fmt.Errorf("failed to read OFX file: %w", err)
	}

	content := string(raw)
	var rows []parsedRow

	// Extract all STMTTRN blocks
	idx := 0
	for {
		start := strings.Index(content[idx:], "<STMTTRN>")
		if start == -1 {
			break
		}
		start += idx
		end := strings.Index(content[start:], "</STMTTRN>")
		if end == -1 {
			break
		}
		end += start + len("</STMTTRN>")

		block := content[start:end]

		date := extractOFXTag(block, "DTPOSTED")
		amount := extractOFXTag(block, "TRNAMT")
		name := extractOFXTag(block, "NAME")
		memo := extractOFXTag(block, "MEMO")
		trnType := extractOFXTag(block, "TRNTYPE")

		desc := name
		if memo != "" && memo != name {
			if desc != "" {
				desc += " | " + memo
			} else {
				desc = memo
			}
		}

		// Parse OFX date: YYYYMMDD or YYYYMMDDHHMMSS
		if len(date) >= 8 {
			date = date[:4] + "-" + date[4:6] + "-" + date[6:8]
		}

		if date != "" && amount != "" {
			rows = append(rows, parsedRow{
				Date:        date,
				Amount:      amount,
				Description: desc,
				Type:        strings.ToLower(trnType),
			})
		}

		idx = end
	}

	if len(rows) == 0 {
		return nil, fmt.Errorf("no transactions found in OFX file — expected <STMTTRN> elements")
	}

	return rows, nil
}

func extractOFXTag(block, tag string) string {
	// Look for <TAG>value or <TAG>value\n
	start := strings.Index(block, "<"+tag+">")
	if start == -1 {
		return ""
	}
	start += len("<" + tag + ">")

	// Value ends at next < or newline
	rest := block[start:]
	endIdx := strings.IndexAny(rest, "<\r\n")
	if endIdx == -1 {
		return strings.TrimSpace(rest)
	}
	return strings.TrimSpace(rest[:endIdx])
}

// parseQIF parses QIF (Quicken Interchange Format) files.
func parseQIF(data io.Reader) ([]parsedRow, error) {
	raw, err := io.ReadAll(data)
	if err != nil {
		return nil, fmt.Errorf("failed to read QIF file: %w", err)
	}

	var rows []parsedRow
	var current parsedRow
	hasData := false

	scanner := bufio.NewScanner(bytes.NewReader(raw))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if len(line) == 0 {
			continue
		}

		switch line[0] {
		case '!':
			// Header/account type, skip
			continue
		case 'D':
			current.Date = line[1:]
			hasData = true
		case 'T', 'U':
			current.Amount = line[1:]
		case 'P':
			current.Description = line[1:]
		case 'M':
			if current.Description == "" {
				current.Description = line[1:]
			}
		case 'N':
			// Check number, can be used as type hint
		case '^':
			// End of record
			if hasData && current.Amount != "" {
				rows = append(rows, current)
			}
			current = parsedRow{}
			hasData = false
		}
	}

	// Handle last record if no trailing ^
	if hasData && current.Amount != "" {
		rows = append(rows, current)
	}

	if len(rows) == 0 {
		return nil, fmt.Errorf("no transactions found in QIF file — expected D/T/P fields")
	}

	return rows, nil
}

func parseDate(s string) time.Time {
	s = strings.TrimSpace(s)
	formats := []string{
		"2006-01-02",
		"01/02/2006",
		"02/01/2006",
		"1/2/2006",
		"2006/01/02",
		"Jan 2, 2006",
		"2 Jan 2006",
		"02-01-2006",
		"01-02-2006",
		"2006.01.02",
		"02.01.2006",
		"2.1.2006",
		"2006-01-02T15:04:05",
		"2006-01-02 15:04:05",
		"01/02/06",
		"1/2/06",
		"02-Jan-2006",
		time.RFC3339,
	}
	for _, f := range formats {
		if t, err := time.Parse(f, s); err == nil {
			return t
		}
	}
	return time.Time{}
}

func parseAmount(s string) (float64, error) {
	s = strings.ReplaceAll(s, " ", "")
	s = strings.ReplaceAll(s, "$", "")
	s = strings.ReplaceAll(s, "€", "")
	s = strings.ReplaceAll(s, "£", "")
	s = strings.ReplaceAll(s, "CHF", "")
	s = strings.ReplaceAll(s, "kr", "")
	s = strings.TrimSpace(s)

	neg := false
	if strings.HasPrefix(s, "(") && strings.HasSuffix(s, ")") {
		neg = true
		s = s[1 : len(s)-1]
	}

	dotIdx := strings.LastIndex(s, ".")
	commaIdx := strings.LastIndex(s, ",")

	if dotIdx >= 0 && commaIdx >= 0 {
		if dotIdx > commaIdx {
			s = strings.ReplaceAll(s, ",", "")
		} else {
			s = strings.ReplaceAll(s, ".", "")
			s = strings.ReplaceAll(s, ",", ".")
		}
	} else if commaIdx >= 0 {
		afterComma := s[commaIdx+1:]
		if len(afterComma) <= 2 {
			s = strings.ReplaceAll(s, ",", ".")
		} else {
			s = strings.ReplaceAll(s, ",", "")
		}
	}

	val, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0, err
	}
	if neg {
		val = -val
	}
	return val, nil
}

func inferTransactionType(typeRaw string, amount float64) models.TransactionType {
	t := strings.ToLower(strings.TrimSpace(typeRaw))
	switch t {
	case "debit", "expense", "payment", "charge":
		return models.TransactionExpense
	case "credit", "income", "refund", "deposit":
		return models.TransactionIncome
	}
	if amount > 0 {
		return models.TransactionIncome
	}
	return models.TransactionExpense
}

func normalizeCurrencyCode(raw string) string {
	c := strings.ToUpper(strings.TrimSpace(raw))
	if len(c) == 3 {
		return c
	}
	return ""
}

func sanitizeCSVValue(value string) string {
	if len(value) == 0 {
		return value
	}
	switch value[0] {
	case '=', '+', '-', '@', '|', '\t', '\r':
		return "'" + value
	}
	return value
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return strings.TrimSpace(v)
		}
	}
	return ""
}

func categoryCacheKey(t models.CategoryType, name string) string {
	return string(t) + "|" + strings.ToLower(strings.TrimSpace(name))
}

func (s *ImportService) resolveOrCreateCategory(
	userID uuid.UUID,
	txType models.TransactionType,
	rawName string,
	cache map[string]uuid.UUID,
) (uuid.UUID, error) {
	catType := models.CategoryExpense
	if txType == models.TransactionIncome {
		catType = models.CategoryIncome
	}

	name := strings.TrimSpace(rawName)
	if name == "" {
		return uuid.Nil, fmt.Errorf("empty category")
	}
	key := categoryCacheKey(catType, name)
	if id, ok := cache[key]; ok {
		return id, nil
	}

	cat := s.buildCategory(userID, name, catType)
	if err := s.categoryRepo.Create(cat); err != nil {
		return uuid.Nil, err
	}
	cache[key] = cat.ID
	return cat.ID, nil
}

func (s *ImportService) resolveOrCreateCategoryTx(
	dbTx *sqlx.Tx,
	userID uuid.UUID,
	txType models.TransactionType,
	rawName string,
	cache map[string]uuid.UUID,
) (uuid.UUID, error) {
	catType := models.CategoryExpense
	if txType == models.TransactionIncome {
		catType = models.CategoryIncome
	}

	name := strings.TrimSpace(rawName)
	if name == "" {
		return uuid.Nil, fmt.Errorf("empty category")
	}
	key := categoryCacheKey(catType, name)
	if id, ok := cache[key]; ok {
		return id, nil
	}

	cat := s.buildCategory(userID, name, catType)
	if err := s.categoryRepo.CreateWithTx(dbTx, cat); err != nil {
		return uuid.Nil, err
	}
	cache[key] = cat.ID
	return cat.ID, nil
}

func (s *ImportService) buildCategory(userID uuid.UUID, name string, catType models.CategoryType) *models.Category {
	icon := "tag"
	color := "#94A3B8"
	switch strings.ToLower(name) {
	case "groceries":
		icon, color = "shopping-cart", "#F43F5E"
	case "food and drink", "restaurant", "restaurants":
		icon, color = "utensils", "#EC4899"
	case "shopping":
		icon, color = "shopping-bag", "#D946EF"
	case "services":
		icon, color = "tool", "#6366F1"
	case "entertainment":
		icon, color = "film", "#A855F7"
	case "payment":
		icon, color = "credit-card", "#3B82F6"
	}

	return &models.Category{
		ID:     uuid.New(),
		UserID: userID,
		Name:   name,
		Type:   catType,
		Icon:   icon,
		Color:  color,
	}
}
