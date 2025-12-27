package service

import "time"

// CertFilter holds the criteria for filtering certificates.
type CertFilter struct {
	AgentID     string
	SearchQuery string
	ValidAfter  *time.Time
	ValidBefore *time.Time
	Limit       int
	Offset      int
	IsTrusted   *bool  // nil=All, true=Trusted, false=Untrusted
	Status      string // ""=All, "ACTIVE", "MISSING"
}

// FilterOption is the function type for the Functional Options pattern.
type FilterOption func(*CertFilter)

func WithAgent(agentID string) FilterOption {
	return func(f *CertFilter) { f.AgentID = agentID }
}

func WithSearch(query string) FilterOption {
	return func(f *CertFilter) { f.SearchQuery = query }
}

func WithPagination(limit, offset int) FilterOption {
	return func(f *CertFilter) {
		if limit > 0 {
			f.Limit = limit
		}
		if offset >= 0 {
			f.Offset = offset
		}
	}
}

func WithExpiryRange(after, before *time.Time) FilterOption {
	return func(f *CertFilter) {
		f.ValidAfter = after
		f.ValidBefore = before
	}
}

// Filter by Trust Status
func WithTrust(isTrusted *bool) FilterOption {
	return func(f *CertFilter) {
		f.IsTrusted = isTrusted
	}
}

// Filter by Presence Status (Active/Missing)
func WithStatus(status string) FilterOption {
	return func(f *CertFilter) {
		f.Status = status
	}
}
