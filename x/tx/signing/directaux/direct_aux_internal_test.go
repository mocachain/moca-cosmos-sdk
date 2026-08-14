package directaux

import "testing"

// TestSameHexAddress exercises sameHexAddress directly, independent of
// GetSignBytes's own test (which can't currently run in this repo — its
// setup hits the unrelated, pre-existing "validator address codec is
// required" failure common to this fork's test infra). The two "garbage
// input" cases mirror direct_aux_test.go's own fixtures ("feePayer" as a
// placeholder address, and the empty-string fee-payer fallback case) to
// make sure this normalization doesn't silently change what that test
// expects once its unrelated blocker is fixed.
func TestSameHexAddress(t *testing.T) {
	tests := []struct {
		name string
		a, b string
		want bool
	}{
		{"identical lowercase", "0xabc1230000000000000000000000000000000d", "0xabc1230000000000000000000000000000000d", true},
		{"same address, different case", "0xABC1230000000000000000000000000000000D", "0xabc1230000000000000000000000000000000d", true},
		{"same address, one missing 0x prefix", "abc1230000000000000000000000000000000d", "0xABC1230000000000000000000000000000000D", true},
		{"different addresses", "0xabc1230000000000000000000000000000000d", "0xdef4560000000000000000000000000000000a", false},
		{"both empty", "", "", true},
		{"empty vs valid non-zero address", "", "0xabc1230000000000000000000000000000000d", false},
		{"both non-hex garbage, identical (matches direct_aux_test.go's placeholder fixture)", "feePayer", "feePayer", true},
		{"both non-hex garbage, different", "feePayer", "notFeePayer", false},
		{"one valid hex, one non-hex garbage", "0xabc1230000000000000000000000000000000d", "feePayer", false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := sameHexAddress(tc.a, tc.b); got != tc.want {
				t.Errorf("sameHexAddress(%q, %q) = %v, want %v", tc.a, tc.b, got, tc.want)
			}
		})
	}
}
