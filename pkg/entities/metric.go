package entities

type Metric struct {
	Key       string
	Value     float64
	Timestamp int64
	Step      int64
}
