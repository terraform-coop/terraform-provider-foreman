package provider

import "testing"

func TestFirstOutOfOrderPair(t *testing.T) {
	cases := []struct {
		name       string
		ids        []string
		wantPrev   string
		wantNext   string
		wantResult bool
	}{
		{"empty", nil, "", "", false},
		{"single", []string{"eth0"}, "", "", false},
		{"already sorted", []string{"eth0", "eth1", "eth2"}, "", "", false},
		{"already sorted, non-numeric suffixes", []string{"ens160", "ens192"}, "", "", false},
		{"out of order", []string{"eth1", "eth0"}, "eth1", "eth0", true},
		{"out of order later in list", []string{"eth0", "eth2", "eth1"}, "eth2", "eth1", true},
		{"duplicate identifiers", []string{"eth0", "eth0"}, "", "", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			prev, next, found := firstOutOfOrderPair(tc.ids)
			if found != tc.wantResult || prev != tc.wantPrev || next != tc.wantNext {
				t.Errorf("firstOutOfOrderPair(%v) = (%q, %q, %v), want (%q, %q, %v)",
					tc.ids, prev, next, found, tc.wantPrev, tc.wantNext, tc.wantResult)
			}
		})
	}
}
