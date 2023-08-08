package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	_ "embed"

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

//go:generate sh -c "printf %s $(git rev-parse HEAD) > .VERSION"
//go:embed .VERSION
var currentCommit string
var updateAvailable = false

func main() {
	localClient = tailscale.LocalClient{}

	localClient.Socket = paths.DefaultTailscaledSocket()
	localClient.UseSocketOnly = true

	ctx = context.Background()

	go checkForUpdates()
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

// reloadDaemon reloads the systray every 15 seconds
func reloadDaemon() {
	for {
		time.Sleep(15 * time.Second)
		reload()
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

	if updateAvailable {
		updateItem := systray.AddMenuItem("Update Available", "Update Available")
		setListener(updateItem, func(interface{}) {
			openUrl("https://github.com/c10udburst/tailscale-systray/")
		}, nil)
		systray.AddSeparator()
	}

	statusItem := systray.AddMenuItem(statusString(status), "Current status of Tailscale")
	statusItem.Disable()

	sent, recv := calculateTraffic(status)
	trafficItem := systray.AddMenuItem(fmt.Sprintf("%s received | %s sent | %d link", fmtByte(recv), fmtByte(sent), len(status.Peers())), "Traffic")
	trafficItem.Disable()

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
	setDeviceList(deviceList, status, prefs)

	systray.AddSeparator()

	preferences := systray.AddMenuItem("Preferences", "Preferences")
	setPreferences(preferences, prefs)

	adminConsole := systray.AddMenuItem("Admin Console", "Admin Console")
	setListener(adminConsole, func(interface{}) {
		openUrl(prefs.AdminPageURL())
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
		name := peerName(node, status)
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

func setDeviceList(root *systray.MenuItem, status *ipnstate.Status, prefs *ipn.Prefs) {
	for _, device := range status.Peer {
		var ip string
		if len(device.TailscaleIPs) > 0 {
			ip = device.TailscaleIPs[0].String()
		}
		name := peerName(device, status)
		item := root.AddSubMenuItem(name, name)
		if !device.Online {
			item.SetIcon(iconOff)
		} else {
			item.SetIcon(iconOn)
		}
		setListener(item, func(ip interface{}) {
			ip = ip.(string)
			openUrl(fmt.Sprintf("%s/%s", prefs.AdminPageURL(), ip))
		}, ip)
	}
}

func checkForUpdates() {
	fmt.Printf("Curr: %s\n", currentCommit)

	var latestCommit string = ""
	resp, err := http.Get("https://api.github.com/repos/C10udburst/tailscale-systray/tags")
	if err != nil {
		return
	}
	defer resp.Body.Close()

	var tags []struct {
		Name   string `json:"name"`
		Commit struct {
			SHA string `json:"sha"`
		} `json:"commit"`
	}
	err = json.NewDecoder(resp.Body).Decode(&tags)
	if err != nil {
		return
	}

	for _, tag := range tags {
		if tag.Name == "latest" {
			latestCommit = tag.Commit.SHA
		}
	}

	if latestCommit == "" {
		return
	}

	fmt.Printf("Curr: %s, Latest: %s\n", currentCommit, latestCommit)

	updateAvailable = currentCommit != latestCommit
}
