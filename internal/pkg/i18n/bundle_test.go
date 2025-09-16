package i18n

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFlattenMap(t *testing.T) {
	t.Run("should flatten a nested map", func(t *testing.T) {
		// Arrange
		data := map[string]any{
			"a": "1",
			"b": map[string]any{
				"c": "2",
				"d": map[string]any{
					"e": "3",
				},
			},
		}
		expected := map[string]string{
			"a":     "1",
			"b.c":   "2",
			"b.d.e": "3",
		}

		// Act
		actual := flattenMap(data)

		// Assert
		assert.Equal(t, expected, actual)
	})

	t.Run("should handle an empty map", func(t *testing.T) {
		// Arrange
		data := make(map[string]any)
		expected := make(map[string]string)

		// Act
		actual := flattenMap(data)

		// Assert
		assert.Equal(t, expected, actual)
	})

	t.Run("should handle a non-nested map", func(t *testing.T) {
		// Arrange
		data := map[string]any{
			"a": "1",
			"b": "2",
		}
		expected := map[string]string{
			"a": "1",
			"b": "2",
		}

		// Act
		actual := flattenMap(data)

		// Assert
		assert.Equal(t, expected, actual)
	})
}

func TestLoad(t *testing.T) {
	t.Run("should load and flatten a valid yaml file", func(t *testing.T) {
		// Arrange
		path := filepath.Join("testdata", "translator", "en-US.yml")

		// Act
		bundle, err := load(path)

		// Assert
		require.NoError(t, err)
		assert.Equal(t, "Hello", bundle["greeting"])
		assert.Equal(t, "Welcome to our application!", bundle["nested.welcome"])
	})

	t.Run("should return an error for a non-existent file", func(t *testing.T) {
		// Arrange
		path := filepath.Join("testdata", "bundle", "non-existent.yml")

		// Act
		_, err := load(path)

		// Assert
		assert.Error(t, err)
	})

	t.Run("should return an error for an invalid yaml file", func(t *testing.T) {
		// Arrange
		path := filepath.Join("testdata", "bundle", "invalid.yml")

		// Act
		_, err := load(path)

		// Assert
		assert.Error(t, err)
	})
}

func TestBundle_Translate(t *testing.T) {
	bundle := Bundle{
		"greeting": "Hello, %s!",
		"farewell": "Goodbye, %s.",
	}

	t.Run("should translate a key with formatting", func(t *testing.T) {
		// Act
		result := bundle.Translate("greeting", "World")

		// Assert
		assert.Equal(t, "Hello, World!", result)
	})

	t.Run("should return the key if it is not found", func(t *testing.T) {
		// Act
		result := bundle.Translate("not_found", "World")

		// Assert
		assert.Equal(t, "not_found", result)
	})
}
