package main

import (
	"fmt"
	"log"
	"os/exec"
	"runtime"

	"tailscale.com/ipn/ipnstate"
	"tailscale.com/util/dnsname"
)

func openUrl(url string) {
	var err error
	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	default:
		err = fmt.Errorf("unsupported platform")
	}
	if err != nil {
		log.Printf("could not open link: %v", err)
	}
}

func peerName(peer *ipnstate.PeerStatus, status *ipnstate.Status) string {
	name := dnsname.TrimSuffix(peer.DNSName, status.CurrentTailnet.MagicDNSSuffix)
	if name == "" {
		return peer.HostName
	}
	return dnsname.SanitizeHostname(name)
}

func statusString(status *ipnstate.Status) string {
	if status.BackendState != "Running" {
		return status.BackendState
	} else {
		var ip = status.Self.DNSName
		if len(status.Self.TailscaleIPs) > 0 {
			ip = status.Self.TailscaleIPs[0].String()
		}
		var msg = fmt.Sprintf("Running: %s (%s)", peerName(status.Self, status), ip)
		if !status.TUN {
			msg += " (no tun)"
		}
		return msg
	}
}

func calculateTraffic(status *ipnstate.Status) (sent int64, recv int64) {
	var sentTotal, recvTotal int64

	for _, node := range status.Peer {
		if node.ShareeNode {
			continue
		}

		sentTotal += node.TxBytes
		recvTotal += node.RxBytes
	}

	return sentTotal, recvTotal
}

func fmtByte(b int64) string {
	const unit = 1000
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB",
		float64(b)/float64(div), "kMGTPE"[exp])
}
