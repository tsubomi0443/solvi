package valueobject

import "fmt"

type SupportStatus int

const (
	SupportStatusPending     SupportStatus = 1
	SupportStatusSupporting  SupportStatus = 2
	SupportStatusDone        SupportStatus = 3
	SupportStatusNotRequired SupportStatus = 4
)

func (s SupportStatus) String() string {
	switch s {
	case SupportStatusPending:
		return "pending"
	case SupportStatusSupporting:
		return "supporting"
	case SupportStatusDone:
		return "done"
	case SupportStatusNotRequired:
		return "not_required"
	default:
		return ""
	}
}

func (s SupportStatus) Int() int { return int(s) }

func ParseSupportStatus(v int) (SupportStatus, error) {
	switch SupportStatus(v) {
	case SupportStatusPending, SupportStatusSupporting, SupportStatusDone, SupportStatusNotRequired:
		return SupportStatus(v), nil
	default:
		return 0, fmt.Errorf("invalid support status: %d", v)
	}
}
