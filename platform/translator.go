package platform

import "github.com/disgoorg/disgo/discord"

type Translator interface {
	// Translate translates a key into the specified locale, using args for formatting.
	Translate(locale discord.Locale, key string, args ...any) string

	// LocalizedMap returns a map of all available localizations for the given key.
	LocalizedMap(key string) map[discord.Locale]string
}
