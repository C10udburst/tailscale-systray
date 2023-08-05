package tailscale

import (
	"fmt"
	"os/exec"

	"github.com/c10udburst/tailscale-systray/options"
)

func Up(options *options.Options) error {
	args, err := options.GetCommand()
	if err != nil {
		return err
	}
	ret, err := exec.Command("tailscale", args...).Output()
	if err != nil {
		return fmt.Errorf("error running tailscale: %w, %s", err, ret)
	}
	return nil
}

func Down() error {
	ret, err := exec.Command("tailscale", "down").Output()
	if err != nil {
		return fmt.Errorf("error running tailscale: %w, %s", err, ret)
	}
	return nil
}
