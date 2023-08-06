package main

import (
	"context"
	_ "embed"
	"fmt"
	"time"

	"fyne.io/systray"
	"github.com/gen2brain/beeep"
	"tailscale.com/client/tailscale"
	"tailscale.com/ipn"
	"tailscale.com/ipn/ipnstate"
	"tailscale.com/paths"
)

var (
	//go:embed icons/on.png
	iconOn []byte
	//go:embed icons/off.png
	iconOff []byte
)

var localClient tailscale.LocalClient
var ctx context.Context

func main() {
	localClient = tailscale.LocalClient{}

	localClient.Socket = paths.DefaultTailscaledSocket()
	localClient.UseSocketOnly = true

	ctx = context.Background()

	go reloadDaemon()
	systray.Run(onReady, onSystrayError)
}

func reload() {
	systray.ResetMenu()
	onReady()
}

func onError(err error) {
	beeep.Notify("Tailscale", err.Error(), "")
}

// reloadDaemon reloads the systray icon every 5 seconds
func reloadDaemon() {
	for {
		time.Sleep(5 * time.Second)
		status, err := localClient.StatusWithoutPeers(ctx)
		if err != nil {
			onError(err)
		} else {
			if status.BackendState == "Running" {
				systray.SetIcon(iconOn)
			} else {
				systray.SetIcon(iconOff)
			}
		}
	}
}

func onSystrayError() {
	onError(fmt.Errorf("an error with systray occurred"))
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

func onReady() {
	systray.SetTitle("Tailscale")
	systray.SetTooltip("Tailscale")

	status, err := localClient.Status(ctx)
	if err != nil {
		onError(err)
		return
	}

	prefs, err := localClient.GetPrefs(ctx)
	if err != nil {
		onError(err)
		return
	}

	statusItem := systray.AddMenuItem(StatusString(status), "Current status of Tailscale")
	statusItem.Disable()

	refresh := systray.AddMenuItem("Refresh", "Refresh this menu")
	setListener(refresh, func(interface{}) {
		reload()
	}, nil)

	systray.AddSeparator()

	connect := systray.AddMenuItem("Connect", "Connect to Tailscale")
	disconnect := systray.AddMenuItem("Disconnect", "Disconnect from Tailscale")
	systray.AddSeparator()

	if status.BackendState == "Running" {
		systray.SetIcon(iconOn)
		connect.Disable()
		setListener(disconnect, func(interface{}) {
			tailscaleDown()
		}, nil)
	} else {
		disconnect.Disable()
		systray.SetIcon(iconOff)
		setListener(connect, func(interface{}) {
			tailscaleUp()
		}, nil)
	}

	var exitNodeFmt = "Exit Node"
	if !prefs.ExitNodeID.IsZero() {
		exitNodeFmt = fmt.Sprintf("Exit Node (%s)", prefs.ExitNodeID)
	}

	exitNode := systray.AddMenuItem(exitNodeFmt, "Exit Node")
	if status.BackendState != "Running" {
		exitNode.Disable()
	}
	setExitNodes(exitNode, status, prefs)

	deviceList := systray.AddMenuItem("Device List", "Device List")
	setDeviceList(deviceList, status)

	systray.AddSeparator()

	preferences := systray.AddMenuItem("Preferences", "Preferences")
	setPreferences(preferences, prefs)

	adminConsole := systray.AddMenuItem("Admin Console", "Admin Console")
	setListener(adminConsole, func(interface{}) {
		OpenUrl(prefs.ControlURL)
	}, nil)
}

func setExitNodes(root *systray.MenuItem, status *ipnstate.Status, prefs *ipn.Prefs) {
	noneItem := root.AddSubMenuItemCheckbox("None", "None", prefs.ExitNodeID.IsZero())

	setListener(noneItem, func(interface{}) {
		setExitNode("")
	}, nil)

	for _, node := range status.Peer {
		if !node.ExitNodeOption {
			continue
		}
		name := PeerName(node, status)
		item := root.AddSubMenuItemCheckbox(name, name, node.ExitNode)
		if node.Online {
			setListener(item, func(data interface{}) {
				node := data.(*ipnstate.PeerStatus)
				setExitNode(node.ID)
				reload()
			}, node)
		} else {
			item.Disable()
		}

	}
}

func setPreferences(root *systray.MenuItem, prefs *ipn.Prefs) {
	allowIncoming := root.AddSubMenuItemCheckbox("Allow Incoming", "Allow Incoming", !prefs.ShieldsUp)
	setListener(allowIncoming, func(interface{}) {
		_, err := localClient.EditPrefs(ctx, &ipn.MaskedPrefs{
			Prefs: ipn.Prefs{
				ShieldsUp: !prefs.ShieldsUp,
			},
			ShieldsUpSet: true,
		})
		if err != nil {
			onError(err)
		}
		beeep.Notify("Tailscale", "Updated settings", "")
		reload()
	}, "")

	acceptRoutes := root.AddSubMenuItemCheckbox("Accept Routes", "Accept Routes", prefs.RouteAll)
	setListener(acceptRoutes, func(interface{}) {
		_, err := localClient.EditPrefs(ctx, &ipn.MaskedPrefs{
			Prefs: ipn.Prefs{
				RouteAll: !prefs.RouteAll,
			},
			RouteAllSet: true,
		})
		if err != nil {
			onError(err)
		}
		beeep.Notify("Tailscale", "Updated settings", "")
		reload()
	}, "")

	acceptDns := root.AddSubMenuItemCheckbox("Accept DNS", "Accept DNS", prefs.CorpDNS)
	setListener(acceptDns, func(interface{}) {
		_, err := localClient.EditPrefs(ctx, &ipn.MaskedPrefs{
			Prefs: ipn.Prefs{
				CorpDNS: !prefs.CorpDNS,
			},
			CorpDNSSet: true,
		})
		if err != nil {
			onError(err)
		}
		beeep.Notify("Tailscale", "Updated settings", "")
		reload()
	}, "")

	exitNodeAllowLan := root.AddSubMenuItemCheckbox("Exit Node Allow Lan", "Exit Node Allow Lan", prefs.ExitNodeAllowLANAccess)
	setListener(exitNodeAllowLan, func(interface{}) {
		_, err := localClient.EditPrefs(ctx, &ipn.MaskedPrefs{
			Prefs: ipn.Prefs{
				ExitNodeAllowLANAccess: !prefs.ExitNodeAllowLANAccess,
			},
			ExitNodeAllowLANAccessSet: true,
		})
		if err != nil {
			onError(err)
		}
		beeep.Notify("Tailscale", "Updated settings", "")
		reload()
	}, "")

	runExitNode := root.AddSubMenuItemCheckbox("Run Exit Node", "Run Exit Node", prefs.AdvertisesExitNode())
	setListener(runExitNode, func(interface{}) {
		copy := prefs.Clone()
		copy.SetAdvertiseExitNode(!prefs.AdvertisesExitNode())
		_, err := localClient.EditPrefs(ctx, &ipn.MaskedPrefs{
			Prefs:              *copy,
			AdvertiseRoutesSet: true,
		})
		if err != nil {
			onError(err)
		}
		beeep.Notify("Tailscale", "Updated settings", "")
		reload()
	}, "")
}

func setDeviceList(root *systray.MenuItem, status *ipnstate.Status) {
	for _, device := range status.Peer {
		var ip string
		if len(device.TailscaleIPs) > 0 {
			ip = device.TailscaleIPs[0].String()
		}
		name := PeerName(device, status)
		item := root.AddSubMenuItem(name, name)
		if !device.Online {
			item.SetIcon(iconOff)
		} else {
			item.SetIcon(iconOn)
		}
		setListener(item, func(ip interface{}) {
			ip = ip.(string)
			OpenUrl(fmt.Sprintf("https://login.tailscale.com/admin/machines/%s", ip))
		}, ip)
	}
}
