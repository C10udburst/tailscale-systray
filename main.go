package main

import (
	"fmt"
	"sync"

	_ "embed"

	"github.com/atotto/clipboard"
	"github.com/c10udburst/tailscale-systray/options"
	"github.com/c10udburst/tailscale-systray/tailscale"
	"github.com/gen2brain/beeep"
	"github.com/getlantern/systray"
)

func main() {
	systray.Run(onReady, nil)
}

var (
	//go:embed icons/on.png
	iconOn []byte
	//go:embed icons/off.png
	iconOff []byte
)

var mu = &sync.Mutex{}

func onReady() {
	mu.Lock()
	systray.SetIcon(iconOff)
	systray.SetTitle("Tailscale")
	systray.SetTooltip("Tailscale")

	options, err := options.ReadOptions()
	if err != nil {
		onErr(err)
	}

	status, err := tailscale.GetStatus()
	if err != nil {
		onErr(err)
	}

	connect := systray.AddMenuItem("Connect", "Connect to Tailscale")
	disconnect := systray.AddMenuItem("Disconnect", "Disconnect from Tailscale")
	systray.AddSeparator()

	setListener(connect, func(interface{}) {
		tailscaleUp()
		connect.Disable()
		disconnect.Enable()
		systray.SetIcon(iconOn)
	}, "")

	setListener(disconnect, func(interface{}) {
		if err := tailscale.Down(); err != nil {
			onErr(err)
		} else {
			beeep.Notify("Tailscale", "Disconnected from Tailscale", "")
			connect.Enable()
			disconnect.Disable()
			systray.SetIcon(iconOff)
		}
	}, "")

	if status.Running {
		systray.SetIcon(iconOn)
		connect.Disable()
	} else {
		disconnect.Disable()
		systray.SetIcon(iconOff)
	}

	exitNode := systray.AddMenuItem("Exit Node", "Exit Node")
	setExitNode(exitNode, options, &status)

	systray.AddSeparator()

	adminConsole := systray.AddMenuItem("Admin Console", "Admin Console")
	setListener(adminConsole, func(interface{}) {
		OpenUrl("https://login.tailscale.com/admin/machines")
	}, "")

	preferences := systray.AddMenuItem("Preferences", "Preferences")
	setPreferences(preferences, options)

	devices := systray.AddMenuItem("Devices", "Devices")
	setDeviceList(devices, &status)

	mu.Unlock()
}

func setExitNode(root *systray.MenuItem, options *options.Options, status *tailscale.Status) {
	noneExitNode := root.AddSubMenuItemCheckbox("None", "None", false)
	var runningExitNode *systray.MenuItem = noneExitNode
	for _, peer := range status.Peers {
		if !peer.ExitNodeOption {
			continue
		}
		peerNode := root.AddSubMenuItemCheckbox(peer.Name, peer.Name, peer.ExitNode)
		setListener(peerNode, func(data interface{}) {
			entry := data.(struct {
				item *systray.MenuItem
				name string
			})
			options.ExitNode = entry.name
			if err := options.Write(); err != nil {
				onErr(err)
			}
			tailscaleUpdate()
			runningExitNode.Uncheck()
			runningExitNode = entry.item
			runningExitNode.Check()
		}, struct {
			item *systray.MenuItem
			name string
		}{
			item: peerNode,
			name: peer.Name,
		})
	}

	runningExitNode.Check()

	setListener(noneExitNode, func(interface{}) {
		options.ExitNode = ""
		if err := options.Write(); err != nil {
			onErr(err)
		}
		tailscaleUpdate()
		runningExitNode.Uncheck()
		runningExitNode = noneExitNode
		runningExitNode.Check()
	}, "")
}

func setPreferences(root *systray.MenuItem, options *options.Options) {
	allowIncoming := root.AddSubMenuItemCheckbox("Allow Incoming", "Allow Incoming", options.AllowIncoming)
	setListener(allowIncoming, func(interface{}) {
		options.AllowIncoming = !options.AllowIncoming
		if err := options.Write(); err != nil {
			onErr(err)
		}
		tailscaleUpdate()
	}, "")

	acceptRoutes := root.AddSubMenuItemCheckbox("Accept Routes", "Accept Routes", options.AcceptRoutes)
	setListener(acceptRoutes, func(interface{}) {
		options.AcceptRoutes = !options.AcceptRoutes
		if err := options.Write(); err != nil {
			onErr(err)
		}
		tailscaleUpdate()
	}, "")

	acceptDns := root.AddSubMenuItemCheckbox("Accept DNS", "Accept DNS", options.AcceptDns)
	setListener(acceptDns, func(interface{}) {
		options.AcceptDns = !options.AcceptDns
		if err := options.Write(); err != nil {
			onErr(err)
		}
		tailscaleUpdate()
	}, "")

	exitNodeAllowLan := root.AddSubMenuItemCheckbox("Exit Node Allow Lan", "Exit Node Allow Lan", options.ExitNodeAllowLan)
	setListener(exitNodeAllowLan, func(interface{}) {
		options.ExitNodeAllowLan = !options.ExitNodeAllowLan
		if err := options.Write(); err != nil {
			onErr(err)
		}
		if options.ExitNode != "" {
			tailscaleUpdate()
		}
	}, "")

	runExitNode := root.AddSubMenuItemCheckbox("Run Exit Node", "Run Exit Node", options.RunExitNode)
	setListener(runExitNode, func(interface{}) {
		options.RunExitNode = !options.RunExitNode
		if err := options.Write(); err != nil {
			onErr(err)
		}
		tailscaleUpdate()
	}, "")
}

func setDeviceList(root *systray.MenuItem, status *tailscale.Status) {
	for _, peer := range status.Peers {
		if peer.Name == "" {
			continue
		}
		item := root.AddSubMenuItem(peer.Name, peer.Name)
		if peer.Online {
			item.SetIcon(iconOn)
		} else {
			item.SetIcon(iconOff)
		}
		setListener(item, func(p interface{}) {
			pr := p.(tailscale.Peer)
			if len(pr.TailscaleIPs) > 0 {
				clipboard.WriteAll(pr.TailscaleIPs[0])
			} else {
				clipboard.WriteAll(pr.DNSName)
			}
			beeep.Notify(pr.Name, fmt.Sprintf("%s\n%+v\nOnline: %t", pr.DNSName, pr.TailscaleIPs, pr.Online), "")
		}, peer)
	}
}

func tailscaleUpdate() {
	status, err := tailscale.GetStatus()
	if err != nil {
		onErr(err)
	}

	if !status.Running {
		return
	}

	tailscaleUp()
}

func tailscaleUp() {

	options, err := options.ReadOptions()
	if err != nil {
		onErr(err)
	}

	if err := tailscale.Up(options); err != nil {
		onErr(err)
	} else {
		beeep.Notify("Tailscale", "Connected to Tailscale", "")
	}
}

func onErr(err error) {
	beeep.Notify("Tailscale Error", err.Error(), "")
}

func setListener(item *systray.MenuItem, listener func(interface{}), data interface{}) {
	go func() {
		for {
			if _, ok := <-item.ClickedCh; !ok {
				break
			}
			listener(data)
		}
	}()
}
