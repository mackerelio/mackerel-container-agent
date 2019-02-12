package spec

import "testing"

func TestGetInterfaces(t *testing.T) {
	ifaces, err := getInterfaces()
	if err != nil {
		t.Errorf("should not raise error: %v", err)
	}

	if len(ifaces) == 0 {
		t.Error("should have at least 1 interface")
	}

	iface := ifaces[0]
	if iface.Name == "" {
		t.Error("interface should have Name")
	}
	if iface.IPAddress == "" {
		t.Error("interface should have IPAddresses")
	}
	if len(iface.IPv4Addresses) == 0 && len(iface.IPv6Addresses) == 0 {
		t.Error("interface should have IPv4Address or IPv6Address")
	}
	if iface.MacAddress == "" {
		t.Errorf("interface should have MacAddress")
	}
}
