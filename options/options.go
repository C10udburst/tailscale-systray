package options

import (
	"encoding/json"
	"fmt"
	"os"
	"os/user"
)

type Options struct {
	AllowIncoming    bool
	AcceptRoutes     bool
	AcceptDns        bool
	RunExitNode      bool
	ExitNode         string
	ExitNodeAllowLan bool
}

const OptionsPath = "options.json"

func ReadOptions() (*Options, error) {
	if _, err := os.Stat(OptionsPath); os.IsNotExist(err) {
		return &Options{
			AllowIncoming: true,
			AcceptRoutes:  true,
			AcceptDns:     true,
			RunExitNode:   false,
			ExitNode:      "",
		}, nil
	}
	dat, err := os.ReadFile(OptionsPath)
	if err != nil {
		return nil, err
	}
	var options Options
	err = json.Unmarshal(dat, &options)
	if err != nil {
		return nil, err
	}
	return &options, nil
}

func (options *Options) Write() error {
	dat, err := json.Marshal(options)
	if err != nil {
		return err
	}
	err = os.WriteFile(OptionsPath, dat, 0644)
	if err != nil {
		return err
	}
	return nil
}

func (o *Options) GetCommand() ([]string, error) {

	u, err := user.Current()
	if err != nil {
		return nil, err
	}

	return []string{
		"up",
		fmt.Sprintf("--accept-dns=%t", o.AcceptDns),
		fmt.Sprintf("--accept-routes=%t", o.AcceptRoutes),
		fmt.Sprintf("--advertise-exit-node=%t", o.RunExitNode),
		fmt.Sprintf("--exit-node=%s", o.ExitNode),
		fmt.Sprintf("--exit-node-allow-lan-access=%t", o.ExitNodeAllowLan),
		fmt.Sprintf("--shields-up=%t", !o.AllowIncoming),
		fmt.Sprintf("--operator=%s", u.Username),
		"--reset",
	}, nil
}
