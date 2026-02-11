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
		"backgroundColor": "#0F172A",
		"paddingAll":      "18px",
		"contents": []interface{}{
			map[string]interface{}{
				"type":   "text",
				"text":   i18n.T(locale, "flex.app_name"),
				"color":  "#CBD5E1",
				"size":   "xs",
				"weight": "bold",
			},
			map[string]interface{}{
				"type":   "text",
				"text":   i18n.Tf(locale, "expense.recorded", map[string]string{"count": fmt.Sprintf("%d", len(expenses)), "amount": formatAmount(totalAmount), "currency": totalCurrency}),
				"color":  "#F8FAFC",
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
				"margin": "md",
			},
		},
	}

	bodyContents := []interface{}{}
	for idx, exp := range expenses {
		if idx > 0 {
			bodyContents = append(bodyContents, map[string]interface{}{
				"type":   "separator",
				"color":  "#E2E8F0",
				"margin": "lg",
			})
		}

		amountText := fmt.Sprintf("%s %s", formatAmount(exp.HomeAmount), exp.HomeCurrency)
		row := map[string]interface{}{
			"type":            "box",
			"layout":          "vertical",
			"paddingAll":      "12px",
			"backgroundColor": "#F8FAFC",
			"cornerRadius":    "12px",
			"margin":          "md",
			"contents": []interface{}{
				map[string]interface{}{
					"type":   "box",
					"layout": "horizontal",
					"contents": []interface{}{
						map[string]interface{}{
							"type":   "text",
							"text":   exp.Description,
							"size":   "sm",
							"color":  "#0F172A",
							"weight": "bold",
							"wrap":   true,
							"flex":   4,
						},
						map[string]interface{}{
							"type":   "text",
							"text":   amountText,
							"size":   "sm",
							"color":  "#0EA5A4",
							"weight": "bold",
							"align":  "end",
							"flex":   3,
						},
					},
				},
			},
		}

		detailParts := []string{}
		if exp.Category != "" {
			detailParts = append(detailParts, exp.Category)
		}
		if !exp.Date.IsZero() {
			detailParts = append(detailParts, exp.Date.Format("2006-01-02"))
		}
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
		if detailText != "" {
			row["contents"] = append(row["contents"].([]interface{}), map[string]interface{}{
				"type":   "text",
				"text":   detailText,
				"size":   "xxs",
				"color":  "#64748B",
				"margin": "xs",
				"wrap":   true,
			})
		}

		if exp.OriginalCurrency != "" && exp.OriginalCurrency != exp.HomeCurrency && exp.OriginalAmount > 0 {
			row["contents"] = append(row["contents"].([]interface{}), map[string]interface{}{
				"type":   "text",
				"text":   fmt.Sprintf("≈ %s %s", formatAmount(exp.OriginalAmount), exp.OriginalCurrency),
				"size":   "xxs",
				"color":  "#64748B",
				"margin": "xs",
			})
		}

		bodyContents = append(bodyContents, row)
	}

	body := map[string]interface{}{
		"type":       "box",
		"layout":     "vertical",
		"paddingAll": "16px",
		"contents":   bodyContents,
	}

	footer := map[string]interface{}{
		"type":            "box",
		"layout":          "vertical",
		"paddingAll":      "14px",
		"backgroundColor": "#F8FAFC",
		"contents": []interface{}{
			map[string]interface{}{
				"type":  "text",
				"text":  i18n.Tf(locale, "expense.count", map[string]string{"count": fmt.Sprintf("%d", len(expenses))}),
				"size":  "xxs",
				"color": "#475569",
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
