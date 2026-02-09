package flex

import (
	"github.com/riverlin/aiexpense/internal/i18n"
)

// BuildReportBubble creates a LINE Flex Message bubble for report link.
func BuildReportBubble(reportURL, locale string) map[string]interface{} {
	header := map[string]interface{}{
		"type":            "box",
		"layout":          "vertical",
		"backgroundColor": "#4F46E5",
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
				"text":   i18n.T(locale, "report.title"),
				"color":  "#FFFFFF",
				"size":   "xl",
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
				"color": "#1E293B",
				"wrap":  true,
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
		"type":       "box",
		"layout":     "vertical",
		"paddingAll": "12px",
		"contents": []interface{}{
			map[string]interface{}{
				"type":  "button",
				"style": "primary",
				"color": "#4F46E5",
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
