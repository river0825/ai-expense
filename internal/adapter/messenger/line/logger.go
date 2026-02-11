package line

import (
	"fmt"
	"log/slog"
)

var lineLogger = slog.Default().With("component", "line_bot")

func maskToken(token string) string {
	if len(token) <= 8 {
		return token
	}
	return fmt.Sprintf("%s...%s", token[:4], token[len(token)-4:])
}

func previewText(text string, max int) string {
	if max <= 0 || len(text) <= max {
		return text
	}
	return text[:max] + "..."
}
