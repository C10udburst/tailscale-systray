package main

import (
	"fmt"

	"github.com/gen2brain/beeep"
	"tailscale.com/ipn"
	"tailscale.com/tailcfg"
)

func tailscaleDown() {
	_, err := localClient.EditPrefs(ctx, &ipn.MaskedPrefs{
		Prefs: ipn.Prefs{
			WantRunning: false,
		},
		WantRunningSet: true,
	})
	if err != nil {
		onError(err)
	}
	beeep.Notify("Tailscale", "Tailscale down", "")
	reload()
}

func tailscaleUp() {
	_, err := localClient.EditPrefs(ctx, &ipn.MaskedPrefs{
		Prefs: ipn.Prefs{
			WantRunning: true,
		},
		WantRunningSet: true,
	})
	if err != nil {
		onError(err)
	}
	beeep.Notify("Tailscale", "Tailscale up", "")
	reload()
}

func setExitNode(node tailcfg.StableNodeID) {
	_, err := localClient.EditPrefs(ctx, &ipn.MaskedPrefs{
		Prefs: ipn.Prefs{
			ExitNodeID: node,
		},
		ExitNodeIDSet: true,
	})
	if err != nil {
		onError(err)
	}
	if node == "" {
		beeep.Notify("Tailscale", "Disabled exit node", "")
	} else {
		beeep.Notify("Tailscale", fmt.Sprint("Set Exit node to", node), "")
	}

	reload()
}
