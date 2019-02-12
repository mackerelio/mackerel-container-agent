package spec

import (
	"net"

	mackerel "github.com/mackerelio/mackerel-client-go"
)

func getInterfaces() ([]mackerel.Interface, error) {

	ifaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}

	values := make([]mackerel.Interface, 0, len(ifaces))

	for _, iface := range ifaces {
		if iface.Flags&net.FlagLoopback != 0 {
			continue
		}
		if iface.HardwareAddr == nil {
			continue
		}

		addrs, err := iface.Addrs()
		if err != nil {
			return nil, err
		}

		ipv4, ipv6 := distinguishAddress(addrs)
		if len(ipv4) == 0 && len(ipv6) == 0 {
			continue
		}

		var ipaddr string
		if len(ipv4) > 0 {
			ipaddr = ipv4[0]
		} else if len(ipv6) > 0 {
			ipaddr = ipv6[0]
		}

		values = append(values, mackerel.Interface{
			Name:          iface.Name,
			MacAddress:    iface.HardwareAddr.String(),
			IPv4Addresses: ipv4,
			IPv6Addresses: ipv6,
			IPAddress:     ipaddr,
		})
	}

	return values, nil
}

func distinguishAddress(addrs []net.Addr) (ipv4 []string, ipv6 []string) {
	for _, addr := range addrs {
		var ip net.IP
		switch addr := addr.(type) {
		case *net.IPNet:
			ip = addr.IP
		case *net.IPAddr:
			ip = addr.IP
		default:
			continue
		}
		if ip == nil {
			continue
		}
		if ip.To4() != nil {
			ipv4 = append(ipv4, ip.String())
		} else if len(ip) == net.IPv6len && ip.To4() == nil {
			ipv6 = append(ipv6, ip.String())
		}
	}
	return
}
