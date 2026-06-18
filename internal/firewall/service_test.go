package firewall

import "testing"

func TestNormalizeRule(t *testing.T) {
	t.Parallel()

	protocol, err := normalizeRule(PortRuleRequest{Port: 443, Protocol: "TCP"})
	if err != nil {
		t.Fatalf("expected valid rule: %v", err)
	}
	if protocol != "tcp" {
		t.Fatalf("expected protocol normalization, got %q", protocol)
	}
}

func TestNormalizeRuleRejectsUnsafeInput(t *testing.T) {
	t.Parallel()

	tests := []PortRuleRequest{
		{Port: 0, Protocol: "tcp"},
		{Port: 65536, Protocol: "tcp"},
		{Port: 443, Protocol: "icmp"},
	}

	for _, test := range tests {
		if _, err := normalizeRule(test); err == nil {
			t.Fatalf("expected invalid rule to fail: %+v", test)
		}
	}
}
