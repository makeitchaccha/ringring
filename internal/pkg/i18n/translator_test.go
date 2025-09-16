package i18n

import (
	"path/filepath"
	"testing"

	"github.com/disgoorg/disgo/discord"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var testdataDir = filepath.Join("testdata", "translator")

const (
	fallbackLocale = "en-US"
)

func TestNewTranslator(t *testing.T) {
	t.Run("should create a new translator successfully", func(t *testing.T) {
		// Arrange & Act
		translator, err := NewTranslator(testdataDir, fallbackLocale)

		// Assert
		require.NoError(t, err)
		require.NotNil(t, translator)
		assert.NotNil(t, translator.fallback)
		assert.Len(t, translator.bundles, 2)
	})

	t.Run("should return an error if the locales directory does not exist", func(t *testing.T) {
		// Arrange & Act
		_, err := NewTranslator("nonexistent", fallbackLocale)

		// Assert
		assert.Error(t, err)
	})

	t.Run("should return an error if the fallback locale is not found", func(t *testing.T) {
		// Arrange & Act
		_, err := NewTranslator(testdataDir, "nonexistent")

		// Assert
		assert.Error(t, err)
	})
}

func TestTranslator_Translate(t *testing.T) {
	// Arrange
	translator, err := NewTranslator(testdataDir, fallbackLocale)
	require.NoError(t, err)

	tests := []struct {
		name     string
		locale   discord.Locale
		key      string
		args     []any
		expected string
	}{
		{
			name:     "should translate in Japanese",
			locale:   "ja",
			key:      "greeting",
			expected: "こんにちは",
		},
		{
			name:     "should translate in English (fallback locale)",
			locale:   "en-US",
			key:      "greeting",
			expected: "Hello",
		},
		{
			name:     "should use fallback locale if key is missing",
			locale:   "ja",
			key:      "only-in-en",
			expected: "This string exists only in the English file.",
		},
		{
			name:     "should return key if it is not found in any bundle",
			locale:   "ja",
			key:      "not_found_key",
			expected: "not_found_key",
		},
		{
			name:     "should use fallback locale for an unknown locale",
			locale:   "fr",
			key:      "greeting",
			expected: "Hello",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			actual := translator.Translate(tt.locale, tt.key, tt.args...)

			// Assert
			assert.Equal(t, tt.expected, actual)
		})
	}
}

func TestTranslator_LocalizedMap(t *testing.T) {
	// Arrange
	translator, err := NewTranslator(testdataDir, fallbackLocale)
	require.NoError(t, err)

	t.Run("should return a map of localized translations for a key", func(t *testing.T) {
		// Act
		localizedMap := translator.LocalizedMap("greeting")

		// Assert
		expected := map[discord.Locale]string{
			discord.LocaleJapanese:  "こんにちは",
			discord.LocaleEnglishUS: "Hello",
		}
		assert.Equal(t, expected, localizedMap)
	})

	t.Run("should return an empty map for a key that does not exist in any bundle", func(t *testing.T) {
		// Act
		localizedMap := translator.LocalizedMap("non_existent_key")

		// Assert
		assert.Empty(t, localizedMap)
	})
}
