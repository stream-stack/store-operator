package discovery

import (
	"context"
	"encoding/json"
	"fmt"
	_ "github.com/Jille/grpc-multi-resolver"
	"github.com/sirupsen/logrus"
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
	"sync"
	"time"
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

	//生成sts地址
	adds := buildPodAddress(broker)
	var hasParitition bool
	var wg sync.WaitGroup

	handler := func(err error, response *http.Response) {
		defer wg.Done()
		if err != nil {
			logrus.Warnf("推送store出现错误,%v", err)
			return
		}
		if response.StatusCode != http.StatusOK {
			logrus.Warnf("http status not ok(200),current:%v", response.StatusCode)
			return
		}
		all, err := ioutil.ReadAll(response.Body)
		if err != nil {
			logrus.Warnf("读取返回body出现错误,%v", err)
			return
		}

		configuration := &protocol.Configuration{}
		err = json.Unmarshal(all, configuration)
		if err != nil {
			logrus.Warnf("反序列化body为Configuration错误,%v", err)
			return
		}
		if len(configuration.Partitions) > 0 {
			hasParitition = true
		}
		logrus.Debugf("推送store后返回值: %s", string(all))
	}
	for _, add := range adds {
		//推送store到dispatcher
		wg.Add(1)
		NewStorePusher(add).push(marshal, handler)
	}
	wg.Wait()
	logrus.Debugf("推送后,是否存在partition:%v", hasParitition)
	if hasParitition {
		logrus.Debugf("exist first Partition,continue")
		return nil
	}
	//分配第一个分片并推送分片
	return allocatePartition(ctx, list.Items, storesetData, broker)
}

const systemBrokerPartition = "_system_broker_partition"

func allocatePartition(ctx context.Context, items []v15.StoreSet, data []protocol.Store, broker *v1.Broker) error {
	logrus.Debugf("开始分配分片,storeset:%d", len(items))
	timeout, cancelFunc := context.WithTimeout(ctx, time.Second*10)
	defer cancelFunc()
	//TODO:根据分片规则分片
	intn := rand.Intn(len(items))
	//set := items[intn]
	s := data[intn]
	//buffer := &bytes.Buffer{}
	marshal, err := json.Marshal(protocol.Partition{
		RangeRegexp: "[0-9]{1,5}",
		Store:       s,
	})
	if err != nil {
		logrus.Errorf("序列化分片数据错误,%v", err)
		return err
	}
	apply, err := store_client.Apply(timeout, s.Uris, &protocol.ApplyRequest{
		StreamName: systemBrokerPartition,
		StreamId:   GetDispatcherStreamId(broker),
		EventId:    "1",
		Data:       marshal,
	})
	if err != nil {
		logrus.Errorf("写入分片出现错误,%v", err)
		return err
	}
	logrus.Debugf("写入分片完成,返回值:%+v", apply)
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
