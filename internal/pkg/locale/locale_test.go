package locale

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoad(t *testing.T) {
	_, err := load("../../../locales/ja.yml")
	assert.NoError(t, err)
}

func TestGetInGoRoutine(t *testing.T) {
	Init("../../../locales")

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		assert.NotNil(t, Get("ja"))
		wg.Done()
	}()
	wg.Wait()

}
