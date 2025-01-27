package locale

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/disgoorg/disgo/discord"
	"gopkg.in/yaml.v3"
)

var (
	fallback Entry

	locales map[discord.Locale]Entry
)

type Entry struct {
	Form struct {
		Settings struct {
			Title struct {
				Guild    string `yaml:"guild"`
				Category string `yaml:"category"`
				Channel  string `yaml:"channel"`
			} `yaml:"title"`
			Description struct {
				Guild    string `yaml:"guild"`
				Category string `yaml:"category"`
				Channel  string `yaml:"channel"`
			} `yaml:"description"`
			Fields struct {
				Notification struct {
					Title  string            `yaml:"title"`
					Update string            `yaml:"update"`
					Values map[string]string `yaml:"values"`
				} `yaml:"notification"`
				NotificationChannel struct {
					Title  string `yaml:"title"`
					Update string `yaml:"update"`
				} `yaml:"notification-channel"`
				ChannelFormat struct {
					Title  string            `yaml:"title"`
					Update string            `yaml:"update"`
					Values map[string]string `yaml:"values"`
				} `yaml:"channel-format"`
				History struct {
					Title  string            `yaml:"title"`
					Update string            `yaml:"update"`
					Values map[string]string `yaml:"values"`
				} `yaml:"history"`
				UsernameFormat struct {
					Title  string            `yaml:"title"`
					Update string            `yaml:"update"`
					Values map[string]string `yaml:"values"`
				} `yaml:"username-format"`
			} `yaml:"fields"`
			Buttons struct {
				ToggleEnability map[string]string `yaml:"toggle-enability"`
				Save            struct {
					Primary       string `yaml:"primary"`
					ConfirmStatus string `yaml:"confirm-status"`
					Confirm       string `yaml:"confirm"`
					Cancel        string `yaml:"cancel"`
				} `yaml:"save"`
				Delete struct {
					Primary       string `yaml:"primary"`
					ConfirmStatus string `yaml:"confirm-status"`
					Confirm       string `yaml:"confirm"`
					Cancel        string `yaml:"cancel"`
				} `yaml:"delete"`
				Discard string `yaml:"discard"`
			} `yaml:"buttons"`
			Validate struct {
				Success string `yaml:"success"`
				Error   struct {
					NoNotificationChannel string `yaml:"no-notification-channel"`
					NoChannelFormat       string `yaml:"no-channel-format"`
					NoPrivacy             string `yaml:"no-privacy"`
					NoUsernameFormat      string `yaml:"no-username-format"`
				} `yaml:"error"`
			} `yaml:"validate"`
			Error struct {
				NotOwner string `yaml:"not-owner"`
			} `yaml:"error"`
		} `yaml:"settings"`
	} `yaml:"form"`
	Command struct {
		Settings struct {
			Description string `yaml:"description"`
			SubCommands struct {
				Guild struct {
					Description string `yaml:"description"`
				} `yaml:"guild"`
				Category struct {
					Description string `yaml:"description"`
					Options     struct {
						Category struct {
							Description string `yaml:"description"`
						} `yaml:"category"`
					} `yaml:"options"`
				} `yaml:"category"`
				Channel struct {
					Description string `yaml:"description"`
					Options     struct {
						Channel struct {
							Description string `yaml:"description"`
						} `yaml:"channel"`
					} `yaml:"options"`
				} `yaml:"channel"`
			} `yaml:"subcommands"`
			Response struct {
				ShowForm string `yaml:"show-form"`
			} `yaml:"response"`
		} `yaml:"settings"`
	} `yaml:"command"`

	Notification struct {
		Common struct {
			StartTime   string `yaml:"start-time"`
			EndTime     string `yaml:"end-time"`
			TimeElapsed string `yaml:"time-elapsed"`
			History     string `yaml:"history"`
			Timeformat  struct {
				Days    string `yaml:"days"`
				Hours   string `yaml:"hours"`
				Minutes string `yaml:"minutes"`
				Seconds string `yaml:"seconds"`
			}
		} `yaml:"common"`
		Ongoing struct {
			Title       string `yaml:"title"`
			Description string `yaml:"description"`
		} `yaml:"ongoing"`
		Ended struct {
			Title       string `yaml:"title"`
			Description string `yaml:"description"`
		} `yaml:"ended"`
	} `yaml:"notification"`
}

func Init(dir string) {
	locales = make(map[discord.Locale]Entry)

	files, err := os.ReadDir(dir)
	if err != nil {
		panic(fmt.Errorf("failed to read directory: %w", err))
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		filename := filepath.Join(dir, file.Name())
		localeEntry, err := load(filename)
		if err != nil {
			panic(fmt.Errorf("failed to load locale: %w", err))
		}

		localeName, _ := strings.CutSuffix(file.Name(), filepath.Ext(file.Name()))

		locales[discord.Locale(localeName)] = localeEntry
	}

	fallback = locales[discord.Locale("en-US")]

}

func load(file string) (Entry, error) {
	f, err := os.OpenFile(file, os.O_RDONLY, 0)

	if err != nil {
		return Entry{}, fmt.Errorf("failed to open file: %w", err)
	}

	defer f.Close()

	var entry Entry
	err = yaml.NewDecoder(f).Decode(&entry)
	if err != nil {
		return Entry{}, fmt.Errorf("failed to decode yaml: %w", err)
	}

	return entry, nil

}

func Get(locale discord.Locale) Entry {
	if entry, ok := locales[locale]; ok {
		return entry
	}

	return fallback
}
