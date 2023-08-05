package tailscale

import (
	"encoding/json"
	"os/exec"

	"tailscale.com/util/dnsname"
)

type Tailnet struct {
	Name            string
	MagicDNSSuffix  string
	MagicDNSEnabled bool
}

type rawPeer struct {
	HostName       string
	DNSName        string
	Online         bool
	ExitNode       bool
	ExitNodeOption bool
	TailscaleIPs   []string
}

type Peer struct {
	rawPeer
	Name string
}

type rawStatus struct {
	BackendState   string
	MagicDNSSuffix string
	Peers          map[string]rawPeer `json:"Peer"`
}

type Status struct {
	BackendState   string
	Running        bool
	MagicDNSSuffix string
	Peers          []Peer
}

func (s rawPeer) toPeer(dnsSuffix string) Peer {
	baseName := dnsname.TrimSuffix(s.DNSName, dnsSuffix)

	var name string
	if baseName == "" {
		name = s.HostName
	} else {
		name = dnsname.SanitizeHostname(baseName)
	}

	return Peer{rawPeer: s, Name: name}
}

func (s *Status) UnmarshalJSON(b []byte) error {
	var raw rawStatus
	err := json.Unmarshal(b, &raw)
	if err != nil {
		return err
	}
	s.BackendState = raw.BackendState
	s.Running = raw.BackendState == "Running"
	s.MagicDNSSuffix = raw.MagicDNSSuffix
	s.Peers = make([]Peer, len(raw.Peers))
	for _, peer := range raw.Peers {
		s.Peers = append(s.Peers, peer.toPeer(raw.MagicDNSSuffix))
	}
	return nil
}

func GetStatus() (Status, error) {
	var status Status
	out, err := exec.Command("tailscale", "status", "--json").Output()
	if err != nil {
		return status, err
	}
	err = json.Unmarshal(out, &status)
	if err != nil {
		return status, err
	}
	return status, nil
}
