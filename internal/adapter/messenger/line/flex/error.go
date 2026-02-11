package flex

import (
	"github.com/riverlin/aiexpense/internal/i18n"
)

// BuildErrorBubble creates a LINE Flex Message bubble for error messages.
func BuildErrorBubble(message, hint, locale string) map[string]interface{} {
	return buildMessageBubble(message, hint, locale, "#B91C1C", i18n.T(locale, "error.title"))
}

// BuildInfoBubble creates a LINE Flex Message bubble for informational messages.
func BuildInfoBubble(message, hint, locale string) map[string]interface{} {
	return buildMessageBubble(message, hint, locale, "#334155", i18n.T(locale, "flex.app_name"))
}

func buildMessageBubble(message, hint, locale, headerColor, title string) map[string]interface{} {
	header := map[string]interface{}{
		"type":            "box",
		"layout":          "vertical",
		"backgroundColor": headerColor,
		"paddingAll":      "18px",
		"contents": []interface{}{
			map[string]interface{}{
				"type":   "text",
				"text":   i18n.T(locale, "flex.app_name"),
				"color":  "#E2E8F0",
				"size":   "xs",
				"weight": "bold",
			},
			map[string]interface{}{
				"type":   "text",
				"text":   title,
				"color":  "#FFFFFF",
				"size":   "lg",
				"weight": "bold",
				"margin": "sm",
			},
		},
	}

	bodyContents := []interface{}{
		map[string]interface{}{
			"type":  "text",
			"text":  message,
			"size":  "sm",
			"color": "#0F172A",
			"wrap":  true,
		},
	}

	if hint != "" {
		bodyContents = append(bodyContents, map[string]interface{}{
			"type":            "box",
			"layout":          "vertical",
			"backgroundColor": "#F8FAFC",
			"cornerRadius":    "10px",
			"paddingAll":      "10px",
			"margin":          "md",
			"contents": []interface{}{
				map[string]interface{}{
					"type":  "text",
					"text":  hint,
					"size":  "xxs",
					"color": "#475569",
					"wrap":  true,
				},
			},
		})
	}

	body := map[string]interface{}{
		"type":       "box",
		"layout":     "vertical",
		"paddingAll": "16px",
		"contents":   bodyContents,
	}

	return map[string]interface{}{
		"type":   "bubble",
		"header": header,
		"body":   body,
	}
}
