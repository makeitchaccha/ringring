package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/makeitchaccha/ringring/internal/app/bot"
	"github.com/makeitchaccha/ringring/internal/pkg/config"
	"github.com/makeitchaccha/ringring/internal/pkg/i18n"
	"gorm.io/gorm"
)

func main() {

	i18n.Init("./locales")

	config, err := config.New("config.yml")

	if err != nil {
		fmt.Fprintln(os.Stderr, "failed to load config:", err)
		os.Exit(1)
	}

	db, err := gorm.Open(config.Dialector, &gorm.Config{})
	if err != nil {
		fmt.Fprintln(os.Stderr, "failed to open database:", err)
		os.Exit(1)
	}

	opts := make([]bot.ConfigOpt, 0)

	if config.Font != nil {
		opts = append(opts, bot.WithFont(config.Font))
	}

	bot, err := bot.New(config.Token, db, opts...)

	if err != nil {
		fmt.Fprintln(os.Stderr, "failed to create bot:", err)
		os.Exit(1)
	}

	if err := bot.Start(context.Background()); err != nil {
		fmt.Fprintln(os.Stderr, "failed to start bot:", err)
		os.Exit(1)
	}

	fmt.Println("Notification Service: Ringring is now running.  Press CTRL-C to exit.")
	s := make(chan os.Signal, 1)
	signal.Notify(s, syscall.SIGINT, syscall.SIGTERM)
	<-s

	bot.Close(context.Background())

}
