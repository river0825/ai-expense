package flex

import (
	"github.com/riverlin/aiexpense/internal/i18n"
)

// BuildErrorBubble creates a LINE Flex Message bubble for error messages.
func BuildErrorBubble(message, hint, locale string) map[string]interface{} {
	return buildMessageBubble(message, hint, locale, "#DC2626", i18n.T(locale, "error.title"))
}

// BuildInfoBubble creates a LINE Flex Message bubble for informational messages.
func BuildInfoBubble(message, hint, locale string) map[string]interface{} {
	return buildMessageBubble(message, hint, locale, "#64748B", i18n.T(locale, "flex.app_name"))
}

func buildMessageBubble(message, hint, locale, headerColor, title string) map[string]interface{} {
	header := map[string]interface{}{
		"type":            "box",
		"layout":          "vertical",
		"backgroundColor": headerColor,
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
				"text":   title,
				"color":  "#FFFFFF",
				"size":   "xl",
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
			"color": "#1E293B",
			"wrap":  true,
		},
	}

	if hint != "" {
		bodyContents = append(bodyContents, map[string]interface{}{
			"type":   "text",
			"text":   hint,
			"size":   "xxs",
			"color":  "#64748B",
			"margin": "md",
			"wrap":   true,
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
