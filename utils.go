package main

import (
	"fmt"
	"log"
	"os/exec"
	"runtime"

	"tailscale.com/ipn/ipnstate"
	"tailscale.com/util/dnsname"
)

func OpenUrl(url string) {
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

func PeerName(peer *ipnstate.PeerStatus, status *ipnstate.Status) string {
	name := dnsname.TrimSuffix(peer.DNSName, status.CurrentTailnet.MagicDNSSuffix)
	if name == "" {
		return peer.HostName
	}
	return dnsname.SanitizeHostname(name)
}
