package i18n

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type Translator struct {
	locales map[string]map[string]string
}

func NewTranslator(localesDir string) (*Translator, error) {
	t := &Translator{
		locales: make(map[string]map[string]string),
	}

	langs := []string{"uz", "ru", "en"}
	for _, lang := range langs {
		path := filepath.Join(localesDir, lang+".json")
		b, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("failed to read locale file %s: %w", path, err)
		}

		var translations map[string]string
		if err := json.Unmarshal(b, &translations); err != nil {
			return nil, fmt.Errorf("failed to parse locale file %s: %w", path, err)
		}

		t.locales[lang] = translations
	}

	return t, nil
}

func (t *Translator) Get(lang, key string) string {
	if translations, ok := t.locales[lang]; ok {
		if val, ok := translations[key]; ok {
			return val
		}
	}
	// Fallback to uzbek if key not found or lang not found
	if uz, ok := t.locales["uz"]; ok {
		if val, ok := uz[key]; ok {
			return val
		}
	}
	return key
}
