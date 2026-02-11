package flex

import (
	"github.com/riverlin/aiexpense/internal/i18n"
)

// BuildReportBubble creates a LINE Flex Message bubble for report link.
func BuildReportBubble(reportURL, locale string) map[string]interface{} {
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
				"text":   i18n.T(locale, "report.title"),
				"color":  "#FFFFFF",
				"size":   "lg",
				"weight": "bold",
				"margin": "sm",
			},
		},
	}

	body := map[string]interface{}{
		"type":       "box",
		"layout":     "vertical",
		"paddingAll": "16px",
		"contents": []interface{}{
			map[string]interface{}{
				"type":  "text",
				"text":  i18n.T(locale, "report.description"),
				"size":  "sm",
				"color": "#0F172A",
				"wrap":  true,
			},
			map[string]interface{}{
				"type":            "box",
				"layout":          "vertical",
				"backgroundColor": "#F8FAFC",
				"cornerRadius":    "12px",
				"paddingAll":      "12px",
				"margin":          "md",
				"contents": []interface{}{
					map[string]interface{}{
						"type":  "text",
						"text":  reportURL,
						"size":  "xxs",
						"color": "#334155",
						"wrap":  true,
					},
				},
			},
			map[string]interface{}{
				"type":   "text",
				"text":   i18n.T(locale, "report.validity"),
				"size":   "xxs",
				"color":  "#64748B",
				"margin": "sm",
			},
		},
	}

	footer := map[string]interface{}{
		"type":            "box",
		"layout":          "vertical",
		"paddingAll":      "14px",
		"backgroundColor": "#F8FAFC",
		"contents": []interface{}{
			map[string]interface{}{
				"type":  "button",
				"style": "primary",
				"color": "#0EA5A4",
				"action": map[string]interface{}{
					"type":  "uri",
					"label": i18n.T(locale, "report.button"),
					"uri":   reportURL,
				},
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
