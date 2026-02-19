package i18n

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestT_ReturnsZhTWByDefault(t *testing.T) {
	result := T("zh-TW", "expense.none")
	assert.Equal(t, "訊息中未偵測到任何支出", result)
}

func TestT_ReturnsEnglish(t *testing.T) {
	result := T("en", "expense.none")
	assert.Equal(t, "No expenses detected in message", result)
}

func TestT_FallsBackToZhTW(t *testing.T) {
	result := T("xx-XX", "expense.none")
	assert.Equal(t, "訊息中未偵測到任何支出", result)
}

func TestT_ReturnsKeyIfNotFound(t *testing.T) {
	result := T("zh-TW", "nonexistent.key")
	assert.Equal(t, "nonexistent.key", result)
}

func TestTf_SubstitutesParams(t *testing.T) {
	result := Tf("zh-TW", "expense.recorded", map[string]string{
		"count":    "3",
		"amount":   "1,250",
		"currency": "TWD",
	})
	assert.Equal(t, "已記錄 3 筆支出，合計：1,250 TWD", result)
}

func TestTf_EnglishWithParams(t *testing.T) {
	result := Tf("en", "expense.recorded", map[string]string{
		"count":    "2",
		"amount":   "500",
		"currency": "USD",
	})
	assert.Equal(t, "Recorded 2 expense(s), total: 500 USD", result)
}

func TestDefaultLocale(t *testing.T) {
	assert.Equal(t, "zh-TW", DefaultLocale())
}

func TestSupportedLocales(t *testing.T) {
	locales := SupportedLocales()
	assert.Contains(t, locales, "zh-TW")
	assert.Contains(t, locales, "en")
}
