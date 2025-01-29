package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/yuyaprgrm/ringring/internal/app/bot"
	"github.com/yuyaprgrm/ringring/internal/pkg/locale"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func main() {

	locale.Init("./locales")

	db, err := gorm.Open(sqlite.Open("ringring.db"), &gorm.Config{})
	if err != nil {
		fmt.Fprintln(os.Stderr, "failed to open database:", err)
		os.Exit(1)
	}

	bot, err := bot.New(os.Getenv("DISCORD_TOKEN"), db)

	if err != nil {
		fmt.Fprintln(os.Stderr, "failed to create bot:", err)
		os.Exit(1)
	}

	if err := bot.Start(context.Background()); err != nil {
		fmt.Fprintln(os.Stderr, "failed to start bot:", err)
		os.Exit(1)
	}

	slog.Info("Test bot is now running.  Press CTRL-C to exit.")
	s := make(chan os.Signal, 1)
	signal.Notify(s, syscall.SIGINT, syscall.SIGTERM)
	<-s

	bot.Close(context.Background())

}
