package services

import (
	"strings"
	"testing"
	"time"

	"moneyvault/internal/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseDate(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		year   int
		month  time.Month
		day    int
		isZero bool
	}{
		{"ISO format", "2024-01-15", 2024, time.January, 15, false},
		{"US format", "01/15/2024", 2024, time.January, 15, false},
		{"US short", "1/2/2024", 2024, time.January, 2, false},
		{"slashed YYYY", "2024/01/15", 2024, time.January, 15, false},
		{"dotted EU", "15.01.2024", 2024, time.January, 15, false},
		{"dotted short", "2.1.2024", 2024, time.January, 2, false},
		{"ISO datetime", "2024-01-15T10:30:00", 2024, time.January, 15, false},
		{"ISO datetime space", "2024-01-15 10:30:00", 2024, time.January, 15, false},
		{"US short year", "01/15/24", 2024, time.January, 15, false},
		{"RFC3339", "2024-01-15T10:30:00Z", 2024, time.January, 15, false},
		{"named month", "Jan 15, 2024", 2024, time.January, 15, false},
		{"named month reversed", "15 Jan 2024", 2024, time.January, 15, false},
		{"dashed named", "15-Jan-2024", 2024, time.January, 15, false},
		{"invalid", "not-a-date", 0, 0, 0, true},
		{"empty", "", 0, 0, 0, true},
		{"garbage", "abc123xyz", 0, 0, 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseDate(tt.input)
			if tt.isZero {
				assert.True(t, result.IsZero())
			} else {
				assert.Equal(t, tt.year, result.Year())
				assert.Equal(t, tt.month, result.Month())
				assert.Equal(t, tt.day, result.Day())
			}
		})
	}
}

func TestParseAmount(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    float64
		wantErr bool
	}{
		{"simple positive", "100.00", 100.00, false},
		{"simple negative", "-50.00", -50.00, false},
		{"with dollar sign", "$100.00", 100.00, false},
		{"with euro sign", "€100.00", 100.00, false},
		{"with pound sign", "£100.00", 100.00, false},
		{"with CHF", "CHF100.00", 100.00, false},
		{"with kr", "kr100.00", 100.00, false},
		{"parentheses negative", "(100.00)", -100.00, false},
		{"comma decimal EU", "1.234,56", 1234.56, false},
		{"comma decimal small", "100,50", 100.50, false},
		{"comma thousands", "1,234.56", 1234.56, false},
		{"comma thousands no dec", "1,234,567", 1234567, false},
		{"with spaces", " 100.00 ", 100.00, false},
		{"spaces within", "1 000.00", 1000.00, false},
		{"zero", "0", 0, false},
		{"invalid", "abc", 0, true},
		{"empty after strip", "", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseAmount(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.InDelta(t, tt.want, got, 0.01)
			}
		})
	}
}

func TestInferTransactionType(t *testing.T) {
	tests := []struct {
		name    string
		typeRaw string
		amount  float64
		want    models.TransactionType
	}{
		{"debit keyword", "debit", 0, models.TransactionExpense},
		{"expense keyword", "expense", 0, models.TransactionExpense},
		{"payment keyword", "Payment", 0, models.TransactionExpense},
		{"charge keyword", "CHARGE", 0, models.TransactionExpense},
		{"credit keyword", "credit", 0, models.TransactionIncome},
		{"income keyword", "income", 0, models.TransactionIncome},
		{"refund keyword", "Refund", 0, models.TransactionIncome},
		{"deposit keyword", "DEPOSIT", 0, models.TransactionIncome},
		{"positive amount", "", 100.00, models.TransactionIncome},
		{"negative amount", "", -50.00, models.TransactionExpense},
		{"zero amount", "", 0, models.TransactionExpense},
		{"unknown type positive", "other", 25.00, models.TransactionIncome},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := inferTransactionType(tt.typeRaw, tt.amount)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestSanitizeCSVValue(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"normal text", "Hello World", "Hello World"},
		{"formula =", "=SUM(A1:A10)", "'=SUM(A1:A10)"},
		{"formula +", "+cmd|' /C calc'!A0", "'+cmd|' /C calc'!A0"},
		{"formula -", "-100", "'-100"},
		{"formula @", "@SUM(A1)", "'@SUM(A1)"},
		{"pipe", "|command", "'|command"},
		{"tab", "\tdata", "'\tdata"},
		{"carriage return", "\rdata", "'\rdata"},
		{"empty", "", ""},
		{"number", "12345", "12345"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := sanitizeCSVValue(tt.input)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestNormalizeCurrencyCode(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"USD", "USD"},
		{"usd", "USD"},
		{"eur", "EUR"},
		{"US", ""},
		{"ABCD", ""},
		{"", ""},
		{" USD ", "USD"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			assert.Equal(t, tt.want, normalizeCurrencyCode(tt.input))
		})
	}
}

func TestFirstNonEmpty(t *testing.T) {
	assert.Equal(t, "hello", firstNonEmpty("hello", "world"))
	assert.Equal(t, "world", firstNonEmpty("", "world"))
	assert.Equal(t, "hello", firstNonEmpty("  ", "hello"))
	assert.Equal(t, "", firstNonEmpty("", "  "))
}

func TestStripBOM(t *testing.T) {
	bom := []byte{0xEF, 0xBB, 0xBF}
	data := append(bom, []byte("hello")...)
	assert.Equal(t, []byte("hello"), stripBOM(data))

	// Without BOM
	assert.Equal(t, []byte("hello"), stripBOM([]byte("hello")))
}

func TestDetectDelimiter(t *testing.T) {
	assert.Equal(t, ',', detectDelimiter([]byte("a,b,c\n1,2,3")))
	assert.Equal(t, ';', detectDelimiter([]byte("a;b;c\n1;2;3")))
	assert.Equal(t, '\t', detectDelimiter([]byte("a\tb\tc\n1\t2\t3")))
	assert.Equal(t, '|', detectDelimiter([]byte("a|b|c\n1|2|3")))
	assert.Equal(t, ',', detectDelimiter([]byte("single")))
}

func TestPreviewCSV(t *testing.T) {
	svc := &ImportService{}

	csv := "Date,Amount,Description\n2024-01-15,100.00,Test payment\n2024-01-16,50.00,Another one\n"
	preview, err := svc.PreviewCSV(strings.NewReader(csv))
	require.NoError(t, err)
	assert.Equal(t, []string{"Date", "Amount", "Description"}, preview.Headers)
	assert.Equal(t, 2, preview.Total)
	assert.Len(t, preview.Rows, 2)
	assert.Equal(t, "100.00", preview.Rows[0].Values["Amount"])
}

func TestPreviewCSV_Empty(t *testing.T) {
	svc := &ImportService{}
	_, err := svc.PreviewCSV(strings.NewReader(""))
	assert.Error(t, err)
}

func TestPreviewCSV_HeadersOnly(t *testing.T) {
	svc := &ImportService{}
	_, err := svc.PreviewCSV(strings.NewReader("Date,Amount\n"))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no data rows")
}

func TestPreviewCSV_SingleColumn(t *testing.T) {
	svc := &ImportService{}
	_, err := svc.PreviewCSV(strings.NewReader("OnlyOne\nvalue\n"))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "column")
}

func TestPreviewCSV_Semicolon(t *testing.T) {
	svc := &ImportService{}
	csv := "Date;Amount;Desc\n2024-01-15;100.00;Test\n"
	preview, err := svc.PreviewCSV(strings.NewReader(csv))
	require.NoError(t, err)
	assert.Equal(t, []string{"Date", "Amount", "Desc"}, preview.Headers)
	assert.Equal(t, "100.00", preview.Rows[0].Values["Amount"])
}

func TestPreviewCSV_MaxPreviewRows(t *testing.T) {
	svc := &ImportService{}
	var b strings.Builder
	b.WriteString("Date,Amount\n")
	for i := 0; i < 20; i++ {
		b.WriteString("2024-01-01,100\n")
	}
	preview, err := svc.PreviewCSV(strings.NewReader(b.String()))
	require.NoError(t, err)
	assert.Equal(t, 20, preview.Total)
	assert.Len(t, preview.Rows, 10) // Max 10 preview rows
}

func TestParseOFX(t *testing.T) {
	ofx := `OFXHEADER:100
<OFX>
<BANKMSGSRSV1>
<STMTTRNRS>
<STMTRS>
<BANKTRANLIST>
<STMTTRN>
<TRNTYPE>DEBIT
<DTPOSTED>20240115120000
<TRNAMT>-50.00
<NAME>Grocery Store
<MEMO>Weekly shopping
</STMTTRN>
<STMTTRN>
<TRNTYPE>CREDIT
<DTPOSTED>20240116
<TRNAMT>1000.00
<NAME>Salary
</STMTTRN>
</BANKTRANLIST>
</STMTRS>
</STMTTRNRS>
</BANKMSGSRSV1>
</OFX>`

	rows, err := parseOFX(strings.NewReader(ofx))
	require.NoError(t, err)
	assert.Len(t, rows, 2)

	assert.Equal(t, "2024-01-15", rows[0].Date)
	assert.Equal(t, "-50.00", rows[0].Amount)
	assert.Equal(t, "Grocery Store | Weekly shopping", rows[0].Description)
	assert.Equal(t, "debit", rows[0].Type)

	assert.Equal(t, "2024-01-16", rows[1].Date)
	assert.Equal(t, "1000.00", rows[1].Amount)
	assert.Equal(t, "Salary", rows[1].Description)
}

func TestParseOFX_NoTransactions(t *testing.T) {
	_, err := parseOFX(strings.NewReader("<OFX></OFX>"))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no transactions")
}

func TestParseQIF(t *testing.T) {
	qif := `!Type:Bank
D01/15/2024
T-50.00
PGrocery Store
^
D01/16/2024
T1000.00
PSalary
^`

	rows, err := parseQIF(strings.NewReader(qif))
	require.NoError(t, err)
	assert.Len(t, rows, 2)

	assert.Equal(t, "01/15/2024", rows[0].Date)
	assert.Equal(t, "-50.00", rows[0].Amount)
	assert.Equal(t, "Grocery Store", rows[0].Description)

	assert.Equal(t, "01/16/2024", rows[1].Date)
	assert.Equal(t, "1000.00", rows[1].Amount)
	assert.Equal(t, "Salary", rows[1].Description)
}

func TestParseQIF_MemoFallback(t *testing.T) {
	qif := `!Type:Bank
D01/15/2024
T-50.00
MFallback memo
^`

	rows, err := parseQIF(strings.NewReader(qif))
	require.NoError(t, err)
	assert.Len(t, rows, 1)
	assert.Equal(t, "Fallback memo", rows[0].Description)
}

func TestParseQIF_NoTrailingCaret(t *testing.T) {
	qif := `!Type:Bank
D01/15/2024
T-50.00
PTest`

	rows, err := parseQIF(strings.NewReader(qif))
	require.NoError(t, err)
	assert.Len(t, rows, 1)
}

func TestParseQIF_NoTransactions(t *testing.T) {
	_, err := parseQIF(strings.NewReader("!Type:Bank\n"))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no transactions")
}

func TestExtractOFXTag(t *testing.T) {
	block := "<TRNTYPE>DEBIT\n<DTPOSTED>20240115\n<TRNAMT>-50.00\n<NAME>Test"

	assert.Equal(t, "DEBIT", extractOFXTag(block, "TRNTYPE"))
	assert.Equal(t, "20240115", extractOFXTag(block, "DTPOSTED"))
	assert.Equal(t, "-50.00", extractOFXTag(block, "TRNAMT"))
	assert.Equal(t, "Test", extractOFXTag(block, "NAME"))
	assert.Equal(t, "", extractOFXTag(block, "MISSING"))
}

func TestCategoryCacheKey(t *testing.T) {
	key := categoryCacheKey(models.CategoryExpense, "Groceries")
	assert.Equal(t, "expense|groceries", key)

	key2 := categoryCacheKey(models.CategoryIncome, "  Salary  ")
	assert.Equal(t, "income|salary", key2)
}
