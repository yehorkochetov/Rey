// Package estimator returns rough monthly USD cost estimates for AWS
// resources flagged by the rey scanners. Prices are intentionally
// conservative us-east-1 list prices — callers that need region-aware or
// savings-plan pricing should layer on top of this baseline.
package estimator

// Resource type identifiers accepted by Estimate. Keep these stable — they
// are compared as strings from scanner code.
const (
	TypeEIP         = "eip"
	TypeEBS         = "ebs"
	TypeSnapshot    = "snapshot"
	TypeRDS         = "rds"
	TypeS3Multipart = "s3-multipart"
	TypeNAT         = "nat"
	TypeVPCEndpoint = "vpc-endpoint"
)

// Estimate returns the monthly USD cost for a resource of the given type.
// For per-GB resources (ebs, snapshot, rds, s3-multipart) sizeGB scales the
// result; flat-priced resources (eip, nat, vpc-endpoint) ignore sizeGB.
// Unknown resource types return 0 rather than an error: scanners should
// surface the resource even when we can't price it.
func Estimate(resourceType string, sizeGB float64) float64 {
	switch resourceType {
	case TypeEIP:
		return 3.60
	case TypeEBS:
		return sizeGB * 0.10
	case TypeSnapshot:
		return sizeGB * 0.095
	case TypeRDS:
		return sizeGB * 0.115
	case TypeS3Multipart:
		return sizeGB * 0.023
	case TypeNAT:
		return 32.40
	case TypeVPCEndpoint:
		return 7.20
	default:
		return 0
	}
}
