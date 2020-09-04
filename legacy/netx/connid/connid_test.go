package connid

import "testing"

func TestIntegrationTCP(t *testing.T) {
	num := Compute("tcp", "1.2.3.4:6789")
	if num != 6789 {
		t.Fatal("unexpected result")
	}
}

func TestIntegrationTCP4(t *testing.T) {
	num := Compute("tcp4", "130.192.91.211:34566")
	if num != 34566 {
		t.Fatal("unexpected result")
	}
}

func TestIntegrationTCP6(t *testing.T) {
	num := Compute("tcp4", "[::1]:4444")
	if num != 4444 {
		t.Fatal("unexpected result")
	}
}

func TestIntegrationUDP(t *testing.T) {
	num := Compute("udp", "1.2.3.4:6789")
	if num != -6789 {
		t.Fatal("unexpected result")
	}
}

func TestIntegrationUDP4(t *testing.T) {
	num := Compute("udp4", "130.192.91.211:34566")
	if num != -34566 {
		t.Fatal("unexpected result")
	}
}

func TestIntegrationUDP6(t *testing.T) {
	num := Compute("udp6", "[::1]:4444")
	if num != -4444 {
		t.Fatal("unexpected result")
	}
}

func TestIntegrationInvalidAddress(t *testing.T) {
	num := Compute("udp6", "[::1]")
	if num != 0 {
		t.Fatal("unexpected result")
	}
}

func TestIntegrationInvalidPort(t *testing.T) {
	num := Compute("udp6", "[::1]:antani")
	if num != 0 {
		t.Fatal("unexpected result")
	}
}

func TestIntegrationNegativePort(t *testing.T) {
	num := Compute("udp6", "[::1]:-1")
	if num != 0 {
		t.Fatal("unexpected result")
	}
}

func TestIntegrationLargePort(t *testing.T) {
	num := Compute("udp6", "[::1]:65536")
	if num != 0 {
		t.Fatal("unexpected result")
	}
}

func TestIntegrationInvalidNetwork(t *testing.T) {
	num := Compute("unix", "[::1]:65531")
	if num != 0 {
		t.Fatal("unexpected result")
	}
}
