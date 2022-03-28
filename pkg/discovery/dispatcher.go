package discovery

import (
	"bytes"
	"context"
	"encoding/binary"
	"encoding/json"
	"fmt"
	_ "github.com/Jille/grpc-multi-resolver"
	v1 "github.com/stream-stack/store-operator/apis/knative/v1"
	v15 "github.com/stream-stack/store-operator/apis/storeset/v1"
	protocol "github.com/stream-stack/store-operator/pkg/proto"
	"github.com/stream-stack/store-operator/pkg/store_client"
	_ "google.golang.org/grpc/health"
	"io/ioutil"
	v13 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/rand"
	"net/http"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"strings"
	"sync"
)

func StartDispatcherStoreSetDiscovery(ctx context.Context, k8sClient client.Client, broker *v1.Broker) error {
	var err error
	list := &v15.StoreSetList{}
	selectorMap, err := v13.LabelSelectorAsMap(broker.Spec.Selector)
	if err != nil {
		return err
	}
	err = k8sClient.List(ctx, list, client.MatchingLabels(selectorMap))
	if err != nil {
		return err
	}
	if len(list.Items) <= 0 {
		return nil
	}
	storesetData := buildStoreData(list)
	marshal, err := json.Marshal(storesetData)
	if err != nil {
		return err
	}
	buffer := bytes.NewBuffer(marshal)

	//生成sts地址
	adds := buildPodAddress(broker)
	var hasParitition bool
	group := sync.WaitGroup{}
	group.Add(len(adds))
	for _, add := range adds {
		//推送store到dispatcher
		NewStorePusher(add).push(buffer, func(err error, response *http.Response) {
			defer group.Done()
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
		})
	}
	group.Wait()
	if hasParitition {
		fmt.Println("exist first Partition,continue")
		return nil
	}
	//分配第一个分片并推送分片
	return allocatePartition(ctx, list.Items, storesetData, broker)
}

const systemBrokerPartition = "_system_broker_partition"

func allocatePartition(ctx context.Context, items []v15.StoreSet, data []protocol.Store, broker *v1.Broker) error {
	//TODO:根据分片规则分片
	intn := rand.Intn(len(items))
	//set := items[intn]
	s := data[intn]
	buffer := &bytes.Buffer{}
	err := binary.Write(buffer, binary.BigEndian, protocol.Partition{
		Begin: "0",
		Store: s,
	})
	if err != nil {
		return err
	}
	apply, err := store_client.Apply(ctx, "dns:///"+strings.Join(s.Uris, ","), &protocol.ApplyRequest{
		StreamName: systemBrokerPartition,
		StreamId:   GetDispatcherStreamId(broker),
		EventId:    "1",
		Data:       buffer.Bytes(),
	})
	if err != nil {
		return err
	}
	fmt.Println("写入第一个分片完成")
	fmt.Println(apply)

	return nil
}

func GetDispatcherStreamId(b *v1.Broker) string {
	return fmt.Sprintf("%s-%s", b.Namespace, b.Name)
}

func GetDispatcherStsName(b *v1.Broker) string {
	return fmt.Sprintf(`%s-dispatcher`, b.Name)
}

func buildStoreData(list *v15.StoreSetList) []protocol.Store {
	data := make([]protocol.Store, len(list.Items))
	for i, item := range list.Items {
		data[i] = protocol.Store{Uris: buildStoreUri(item), Name: item.Name, Namespace: item.Namespace}
	}
	return data
}

func buildStoreUri(item v15.StoreSet) []string {
	replicas := *item.Spec.Store.Replicas
	addrs := make([]string, replicas)
	var i int32
	for ; i < replicas; i++ {
		//TODO:重构名称的生成,应该和模板统一,使用template的自定义函数
		addrs[i] = fmt.Sprintf(`%s-%d.%s.%s`, item.Name, i, item.Status.StoreStatus.ServiceName, item.Namespace)
	}
	return addrs
}

func buildPodAddress(broker *v1.Broker) []string {
	replicas := broker.Spec.Dispatcher.Replicas
	addrs := make([]string, replicas)
	var i int32
	for ; i < replicas; i++ {
		addrs[i] = fmt.Sprintf(`%s-%d.%s.%s`, GetDispatcherStsName(broker), i, broker.Status.Dispatcher.SvcName, broker.GetNamespace())
	}
	return addrs
}
