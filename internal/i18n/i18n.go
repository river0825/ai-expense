package i18n

import (
	"embed"
	"encoding/json"
	"strings"
	"sync"
)

//go:embed locales/*.json
var localeFS embed.FS

const defaultLocale = "zh-TW"

var (
	translations map[string]map[string]string
	once         sync.Once
)

func load() {
	once.Do(func() {
		translations = make(map[string]map[string]string)
		entries, err := localeFS.ReadDir("locales")
		if err != nil {
			return
		}
		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}
			name := entry.Name()
			if !strings.HasSuffix(name, ".json") {
				continue
			}
			locale := strings.TrimSuffix(name, ".json")
			data, err := localeFS.ReadFile("locales/" + name)
			if err != nil {
				continue
			}
			var m map[string]string
			if err := json.Unmarshal(data, &m); err != nil {
				continue
			}
			translations[locale] = m
		}
	})
}

func T(locale, key string) string {
	load()
	if m, ok := translations[locale]; ok {
		if v, ok := m[key]; ok {
			return v
		}
	}
	if locale != defaultLocale {
		if m, ok := translations[defaultLocale]; ok {
			if v, ok := m[key]; ok {
				return v
			}
		}
	}
	return key
}

func Tf(locale, key string, params map[string]string) string {
	tmpl := T(locale, key)
	if len(params) == 0 {
		return tmpl
	}
	oldNew := make([]string, 0, len(params)*2)
	for k, v := range params {
		oldNew = append(oldNew, "{"+k+"}", v)
	}
	return strings.NewReplacer(oldNew...).Replace(tmpl)
}

func SupportedLocales() []string {
	load()
	locales := make([]string, 0, len(translations))
	for k := range translations {
		locales = append(locales, k)
	}
	return locales
}

func DefaultLocale() string {
	return defaultLocale
}
