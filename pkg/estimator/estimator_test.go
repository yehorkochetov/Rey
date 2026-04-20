package estimator

import "testing"

func TestEstimate(t *testing.T) {
	cases := []struct {
		name         string
		resourceType string
		sizeGB       float64
		want         float64
	}{
		{"EIP flat price", TypeEIP, 0, 3.60},
		{"EIP ignores size", TypeEIP, 1000, 3.60},
		{"EBS 100GB", TypeEBS, 100, 10.00},
		{"EBS 1GB", TypeEBS, 1, 0.10},
		{"Snapshot 100GB", TypeSnapshot, 100, 9.50},
		{"RDS 100GB", TypeRDS, 100, 11.50},
		{"S3 multipart 10GB", TypeS3Multipart, 10, 0.23},
		{"NAT gateway", TypeNAT, 0, 32.40},
		{"VPC endpoint", TypeVPCEndpoint, 0, 7.20},
		{"Unknown type", "not-a-real-type", 100, 0},
		{"Empty type", "", 100, 0},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := Estimate(c.resourceType, c.sizeGB)
			if !almostEqual(got, c.want) {
				t.Errorf("Estimate(%q, %v) = %v, want %v", c.resourceType, c.sizeGB, got, c.want)
			}
		})
	}
}

// almostEqual tolerates the tiny float drift produced by multiplying
// non-terminating binary fractions like 0.023 * 10 = 0.22999999999999998.
func almostEqual(a, b float64) bool {
	const epsilon = 1e-9
	d := a - b
	if d < 0 {
		d = -d
	}
	return d < epsilon
}
