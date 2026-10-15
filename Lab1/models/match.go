package models

import (
	"fmt"
	"time"
)

type Match struct {
	Guest       *Team
	Host        *Team
	GuestPoints int
	HostPoints  int
	Date        time.Time
}

func (m *Match) String() string {
	score := "vs"
	if m.HostPoints > 0 || m.GuestPoints > 0 {
		score = fmt.Sprintf("%d : %d", m.HostPoints, m.GuestPoints)
	}
	return fmt.Sprintf("[%s] %s %s %s",
		m.Date.Format("2006-01-02 15:04"),
		m.Host.Name,
		score,
		m.Guest.Name,
	)
}
