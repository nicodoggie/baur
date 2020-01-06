package term

import (
	"fmt"
	"time"
)

func SecondDuration(duration time.Duration) string {
	return fmt.Sprintf("%.3f", duration.Seconds())
}
