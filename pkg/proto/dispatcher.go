package protocol

type Store struct {
	Name      string   `json:"name"`
	Namespace string   `json:"namespace"`
	Uris      []string `json:"uris"`
}

type Partition struct {
	RangeRegexp string `json:"rangeRegexp"`
	Store       Store  `json:"store"`
}

type Configuration struct {
	Stores     []Store     `json:"stores"`
	Partitions []Partition `json:"partitions"`
	MaxEventId string      `json:"max_event_id"`
}
