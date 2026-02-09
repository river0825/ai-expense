package flex

import (
	"fmt"
	"time"

	"github.com/riverlin/aiexpense/internal/i18n"
)

// ExpenseData holds the data needed to render a single expense in a Flex Message
type ExpenseData struct {
	Description      string
	HomeAmount       float64
	HomeCurrency     string
	OriginalAmount   float64
	OriginalCurrency string
	Category         string
	Account          string
	Date             time.Time
}

// BuildExpenseBubble creates a LINE Flex Message bubble for expense confirmation.
func BuildExpenseBubble(expenses []ExpenseData, totalAmount float64, totalCurrency, locale string) map[string]interface{} {
	header := map[string]interface{}{
		"type":            "box",
		"layout":          "vertical",
		"backgroundColor": "#059669",
		"paddingAll":      "16px",
		"contents": []interface{}{
			map[string]interface{}{
				"type":  "text",
				"text":  i18n.T(locale, "flex.app_name"),
				"color": "#FFFFFF",
				"size":  "xs",
			},
			map[string]interface{}{
				"type":   "text",
				"text":   i18n.Tf(locale, "expense.recorded", map[string]string{"count": fmt.Sprintf("%d", len(expenses)), "amount": formatAmount(totalAmount), "currency": totalCurrency}),
				"color":  "#FFFFFF",
				"size":   "sm",
				"margin": "sm",
				"wrap":   true,
			},
			map[string]interface{}{
				"type":   "text",
				"text":   fmt.Sprintf("%s %s %s", i18n.T(locale, "flex.total"), formatAmount(totalAmount), totalCurrency),
				"color":  "#FFFFFF",
				"size":   "xxl",
				"weight": "bold",
				"margin": "sm",
			},
		},
	}

	bodyContents := []interface{}{}
	for idx, exp := range expenses {
		if idx > 0 {
			bodyContents = append(bodyContents, map[string]interface{}{
				"type":   "separator",
				"color":  "#E2E8F0",
				"margin": "md",
			})
		}

		amountText := fmt.Sprintf("%s %s", formatAmount(exp.HomeAmount), exp.HomeCurrency)
		row := map[string]interface{}{
			"type":   "box",
			"layout": "horizontal",
			"margin": "md",
			"contents": []interface{}{
				map[string]interface{}{
					"type":   "text",
					"text":   exp.Description,
					"size":   "sm",
					"color":  "#1E293B",
					"weight": "bold",
					"flex":   4,
				},
				map[string]interface{}{
					"type":   "text",
					"text":   amountText,
					"size":   "sm",
					"color":  "#059669",
					"weight": "bold",
					"align":  "end",
					"flex":   3,
				},
			},
		}
		bodyContents = append(bodyContents, row)

		detailParts := []string{}
		if exp.Category != "" {
			detailParts = append(detailParts, exp.Category)
		}
		detailParts = append(detailParts, exp.Date.Format("2006-01-02"))
		if exp.Account != "" {
			detailParts = append(detailParts, exp.Account)
		}
		detailText := ""
		for i, p := range detailParts {
			if i > 0 {
				detailText += " · "
			}
			detailText += p
		}
		detail := map[string]interface{}{
			"type":   "text",
			"text":   detailText,
			"size":   "xxs",
			"color":  "#64748B",
			"margin": "xs",
		}
		bodyContents = append(bodyContents, detail)

		if exp.OriginalCurrency != "" && exp.OriginalCurrency != exp.HomeCurrency && exp.OriginalAmount > 0 {
			converted := map[string]interface{}{
				"type":   "text",
				"text":   fmt.Sprintf("≈ %s %s", formatAmount(exp.OriginalAmount), exp.OriginalCurrency),
				"size":   "xxs",
				"color":  "#64748B",
				"margin": "xs",
			}
			bodyContents = append(bodyContents, converted)
		}
	}

	body := map[string]interface{}{
		"type":       "box",
		"layout":     "vertical",
		"paddingAll": "16px",
		"contents":   bodyContents,
	}

	footer := map[string]interface{}{
		"type":       "box",
		"layout":     "vertical",
		"paddingAll": "12px",
		"contents": []interface{}{
			map[string]interface{}{
				"type":  "text",
				"text":  i18n.Tf(locale, "expense.count", map[string]string{"count": fmt.Sprintf("%d", len(expenses))}),
				"size":  "xxs",
				"color": "#64748B",
				"align": "center",
			},
		},
	}

	return map[string]interface{}{
		"type":   "bubble",
		"header": header,
		"body":   body,
		"footer": footer,
	}
}

func formatAmount(amount float64) string {
	if amount == float64(int64(amount)) {
		return fmt.Sprintf("%d", int64(amount))
	}
	return fmt.Sprintf("%.2f", amount)
}
