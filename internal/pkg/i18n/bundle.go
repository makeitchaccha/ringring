package i18n

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Bundle map[string]string

func (b Bundle) Translate(key string, args ...any) string {
	if translation, ok := b[key]; ok {
		return fmt.Sprintf(translation, args...)
	}
	return key
}

func flattenMap(data map[string]any) map[string]string {
	result := make(map[string]string)
	internalFlattenMap(data, "", result)
	return result
}

func internalFlattenMap(data map[string]any, prefix string, result map[string]string) {
	for key, value := range data {
		fullKey := key
		if prefix != "" {
			fullKey = prefix + "." + key
		}

		switch v := value.(type) {
		case map[string]any:
			internalFlattenMap(v, fullKey, result)
		case string:
			result[fullKey] = v
		default:
			// Ignore unsupported types
		}
	}
}

func load(file string) (Bundle, error) {
	f, err := os.OpenFile(file, os.O_RDONLY, 0)

	if err != nil {
		return Bundle{}, fmt.Errorf("failed to open file: %w", err)
	}

	defer f.Close()

	var data map[string]any

	if err := yaml.NewDecoder(f).Decode(&data); err != nil {
		return Bundle{}, fmt.Errorf("failed to decode yaml: %w", err)
	}

	entry := flattenMap(data)

	return entry, nil
}
