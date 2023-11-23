package ulog

import (
	"fmt"
	"time"
)

func (e entry) String() string {
	return fmt.Sprintf("time=%s level=%s string=%s", e.Time.Format(time.RFC3339Nano), e.Level, e.Message)
}
