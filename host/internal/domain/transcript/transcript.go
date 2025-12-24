package transcript

import (
	"time"
)

type Chunk struct {
	Text      string
	Timestamp time.Time
}
