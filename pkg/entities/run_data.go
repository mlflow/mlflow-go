package entities

type RunData struct {
	Tags    []*RunTag
	Params  []*Param
	Metrics []*Metric
}
