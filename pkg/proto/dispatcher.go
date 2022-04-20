package proto

import (
	"io"
	"io/ioutil"
	"k8s.io/apimachinery/pkg/util/json"
)

type Statistics struct {
	Uris              []string `json:"uris,omitempty"`
	PartitionCount    uint64   `json:"partitionCount,omitempty"`
	MaxPartitionBegin uint64   `json:"maxPartitionBegin,omitempty"`
	MaxEvent          uint64   `json:"maxEvent,omitempty"`
	TotalDataSize     uint64   `json:"totalDataSize,omitempty"`
}

func ParseStatistics(response io.ReadCloser, out *Statistics) error {
	//read all response body , json marshal to Statistics
	all, err := ioutil.ReadAll(response)
	if err != nil {
		return err
	}
	return json.Unmarshal(all, out)
}

type PartitionAllocator interface {
	AllocatePartition(statistics *Statistics, sets []*StoreSet) (*StoreSet, uint64, error)
}
