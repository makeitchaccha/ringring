package i18n

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/disgoorg/disgo/discord"
)

type Translator struct {
	fallback Bundle
	bundles  map[string]Bundle
}

func NewTranslator(localesDir string, fallbackLocale string) (*Translator, error) {
	fileEntries, err := os.ReadDir(localesDir)

	if err != nil {
		return nil, fmt.Errorf("failed to read locales directory: %w", err)
	}

	bundles := make(map[string]Bundle)
	for _, entry := range fileEntries {
		if entry.IsDir() {
			// Skip directories
			continue
		}

		localeName, _ := strings.CutSuffix(entry.Name(), filepath.Ext(entry.Name()))
		bundle, err := load(filepath.Join(localesDir, entry.Name()))
		if err != nil {
			return nil, fmt.Errorf("failed to load locale %s: %w", localeName, err)
		}

		bundles[localeName] = bundle
	}

	fallback, ok := bundles[fallbackLocale]
	if !ok {
		return nil, fmt.Errorf("fallback locale %s not found", fallbackLocale)
	}

	return &Translator{
		fallback: fallback,
		bundles:  bundles,
	}, nil
}

// Translate is the main translation method.
// It gets the translation for the given locale and key, with fallback support.
func (t *Translator) Translate(locale discord.Locale, key string, args ...any) string {
	// Try to get the bundle for the requested locale
	if bundle, ok := t.bundles[locale.String()]; ok {
		// If the bundle is found, try to get the translation
		if translation, ok := bundle[key]; ok {
			return fmt.Sprintf(translation, args...)
		}
	}

	// If the translation is not found in the requested locale, use the fallback bundle
	if translation, ok := t.fallback[key]; ok {
		return fmt.Sprintf(translation, args...)
	}

	// If the key is not found anywhere, return the key itself
	return key
}

// LocalizedMap returns a map of all translations for a single key.
// This is used for setting localized command descriptions on Discord.
func (t *Translator) LocalizedMap(key string, args ...any) map[discord.Locale]string {
	localized := make(map[discord.Locale]string)
	for localeStr := range t.bundles {
		locale := discord.Locale(localeStr)
		// The Translate method already handles fallback, so we just call it
		translation := t.Translate(locale, key, args...)

		// Add to map only if a translation was found (i.e., it's not the key itself)
		if translation != key {
			localized[locale] = translation
		}
	}
	return localized
}

// Bundle returns the entire translation map for a specific locale, without fallback.
func (t *Translator) Bundle(locale string) (Bundle, bool) {
	bundle, ok := t.bundles[locale]
	return bundle, ok
}
