// ref: call-notify
package visualizer

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
)

var (
	exectablePath string
)

func Init(path string) error {
	cmd := exec.Command(path, "-v")
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to initialize visualizer: %w", err)
	}

	fullVersion := out.String()
	if !strings.HasPrefix(fullVersion, "timeliner by yuyaprgrm: ") {
		return fmt.Errorf("invalid visualizer: %s", fullVersion)
	}

	exectablePath = path
	return nil
}

func Generate(request Request, output string) error {
	buf, err := json.Marshal(request)
	if err != nil {
		return err
	}

	cmd := exec.Command(exectablePath, "-i", "-", "-o", output)
	cmd.Stdin = bytes.NewReader(buf)
	if err := cmd.Run(); err != nil {
		return err
	}

	return nil
}
