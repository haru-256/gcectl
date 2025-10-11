package enums

import "github.com/haru-256/gcectl/pkg/log"

type Status int

// ステップ2: const と iota で定数を定義
const (
	StatusUnknown Status = iota
	StatusRunning
	StatusTerminated
)

func StatusFromString(status string) Status {
	switch status {
	case "RUNNING":
		return StatusRunning
	case "TERMINATED":
		return StatusTerminated
	default:
		log.Logger.Warnf("Unknown status: %s", status)
		return StatusUnknown
	}
}

func (s Status) String() string {
	switch s {
	case StatusRunning:
		return "RUNNING"
	case StatusTerminated:
		return "TERMINATED"
	default:
		return "UNKNOWN"
	}
}

func (s Status) Render() string {
	switch s {
	case StatusRunning:
		return "🟢(RUNNING)"
	case StatusTerminated:
		return "🔴(TERMINATED)"
	default:
		return "UNKNOWN"
	}
}
