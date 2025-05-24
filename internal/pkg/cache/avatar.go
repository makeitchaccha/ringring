package cache

import (
	"fmt"
	"image"
	"net/http"
	"os"
	"time"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/rest"
	"github.com/disgoorg/snowflake/v2"
	"github.com/makeitchaccha/ringring/pkg/extstd"
	"github.com/nfnt/resize"
	"golang.org/x/image/webp"
)

var avatarCache = make(map[snowflake.ID]extstd.Cache[image.Image])

func GetAvatar(rest rest.Rest, id snowflake.ID) (image.Image, error) {
	if avatar, ok := avatarCache[id]; ok && avatar.Valid() {
		return avatar.Unwrap(), nil
	}

	return cacheAvatar(rest, id)
}

func cacheAvatar(rest rest.Rest, id snowflake.ID) (image.Image, error) {
	user, err := rest.GetUser(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	resp, err := http.Get(user.EffectiveAvatarURL(discord.WithSize(64), discord.WithFormat(discord.FileFormatWebP)))
	if err != nil {
		return nil, fmt.Errorf("failed to get avatar: %w", err)
	}
	avatar, err := webp.Decode(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to decode avatar: %w", err)
	}

	if avatar.Bounds().Dx() != 64 || avatar.Bounds().Dy() != 64 {
		fmt.Fprintln(os.Stderr, "avatar is not 64x64")
		// try to resize the image
		avatar = resize.Resize(64, 64, avatar, resize.Lanczos3)
	}

	avatarCache[id] = extstd.NewCache(avatar, 1*time.Hour)
	return avatar, nil
}
