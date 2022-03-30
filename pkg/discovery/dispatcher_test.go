package discovery

import (
	"context"
	"encoding/json"
	"fmt"
	v1 "github.com/stream-stack/store-operator/apis/knative/v1"
	v15 "github.com/stream-stack/store-operator/apis/storeset/v1"
	protocol "github.com/stream-stack/store-operator/pkg/proto"
	"io/ioutil"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"net/http"
	"sync"
	"testing"
)

func TestStartDispatcherStoreSetDiscovery(t *testing.T) {
	adds := []string{"www.baidu.com"}
	marshal := []byte{}

	var hasParitition bool
	var wg sync.WaitGroup
	handler := func(err error, response *http.Response) {
		defer wg.Done()
		if err != nil {
			fmt.Println(err)
			return
		}
		if response.StatusCode != http.StatusOK {
			fmt.Println(fmt.Errorf("http status not ok(200),current:%v", response.StatusCode))
		}
		all, err := ioutil.ReadAll(response.Body)
		if err != nil {
			fmt.Println(err)
			return
		}

		configuration := &protocol.Configuration{}
		err = json.Unmarshal(all, configuration)
		if err != nil {
			fmt.Println(err)
			return
		}
		if len(configuration.Partitions) > 0 {
			hasParitition = true
		}
		fmt.Println("推送store后返回值: ", string(all))
	}
	for _, add := range adds {
		//推送store到dispatcher
		wg.Add(1)
		NewStorePusher(add).push(marshal, handler)
	}
	wg.Wait()
	fmt.Println(hasParitition)
}

func TestAllocatePartition(t *testing.T) {
	stores := []v15.StoreSet{{
		TypeMeta:   metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{},
		Spec:       v15.StoreSetSpec{},
		Status:     v15.StoreSetStatus{},
	}}
	datas := []protocol.Store{{
		Name:      "1",
		Namespace: "1",
		Uris:      []string{"www.baidu.com"},
	}}
	broker := &v1.Broker{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "b1",
			Namespace: "b-namespaces",
		},
	}
	err := allocatePartition(context.TODO(), stores, datas, broker)
	if err != nil {
		fmt.Println(err)
	}
}
