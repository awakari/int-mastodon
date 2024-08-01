package model

type SearchType int

const (
	SearchTypeStatuses SearchType = iota
	SearchTypeAccounts
)

func (typ SearchType) String() string {
	return []string{"statuses", "accounts"}[typ]
}
