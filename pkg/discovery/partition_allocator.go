package discovery

import (
	"context"
	"fmt"
	pp "github.com/golang/protobuf/proto"
	"github.com/sirupsen/logrus"
	"github.com/stream-stack/common/protocol/dispatcher"
	"github.com/stream-stack/common/protocol/operator"
	"github.com/stream-stack/common/protocol/store"
	v12 "github.com/stream-stack/store-operator/apis/knative/v1"
	v13 "github.com/stream-stack/store-operator/apis/storeset/v1"
	"github.com/stream-stack/store-operator/pkg/store_client"
	v1 "k8s.io/api/core/v1"
	v14 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"math/rand"
	"net/http"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"time"
)

var AllocateRequestCh = make(chan v12.Broker, 1)

func StartPartitionAllocator(ctx context.Context, client client.Client) {
	ticker := time.NewTicker(time.Minute * 5)
	defer ticker.Stop()
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				logrus.Info("Start partition allocator")
				//list broker
				brokerList := &v12.BrokerList{}
				if err := client.List(ctx, brokerList); err != nil {
					logrus.Errorf("list broker error: %v", err)
					continue
				}
				//loop broker list
				for _, item := range brokerList.Items {
					handlerRequest(ctx, client, item)
				}
			case request := <-AllocateRequestCh:
				handlerRequest(ctx, client, request)
			}
		}
	}()
}

const statisticsUriFormat = `http://%s:%d/statistics`
const defaultPort = 8080
const defaultTimeOut = time.Second * 5
const systemBrokerPartition = "_system_broker_partition"

func handlerRequest(ctx context.Context, c client.Client, broker v12.Broker) {
	pod, err := getMatchPod(ctx, c, broker)
	if err != nil {
		return
	}
	statistics, err := getStatistics(pod)
	if err != nil {
		return
	}
	//立即分配第一个分片
	storeSets, err := getStoreSetData(ctx, c, broker)
	if err != nil {
		return
	}
	if len(storeSets) == 0 {
		return
	}
	//分配新的分片
	partition, begin, err := broker.Spec.Partition.AllocatePartition(statistics, storeSets)
	if err != nil {
		logrus.Errorf("allocate partition for broker %s/%s error: %v", broker.Namespace, broker.Name, err)
		return
	}
	if partition == nil {
		logrus.Debugf("no partition need to allocate for broker %s/%s", broker.Namespace, broker.Name)
		return
	}
	writePartition(ctx, partition, broker, statistics.PartitionCount+1, begin)
}

func writePartition(ctx context.Context, set *operator.StoreSet, broker v12.Broker, i uint64, begin uint64) {
	bytes, err := pp.Marshal(&operator.Partition{
		Begin: begin,
		Store: set,
	})
	if err != nil {
		logrus.Errorf("protobuf marshal partition error,%v", err)
		return
	}
	apply, err := store_client.Apply(ctx, set.Uris, &store.ApplyRequest{
		StreamName: systemBrokerPartition,
		StreamId:   GetStreamName(&broker),
		EventId:    i,
		Data:       bytes,
	})
	if err != nil {
		logrus.Errorf("write partition for broker %s/%s error,%v", broker.Namespace, broker.Name, err)
		return
	}
	logrus.Debugf("write partition for broker %s/%s success,result:%+v", broker.Namespace, broker.Name, apply)
	return
}

func getStoreSetData(ctx context.Context, c client.Client, broker v12.Broker) ([]*operator.StoreSet, error) {
	list := &v13.StoreSetList{}
	selectorMap, err := v14.LabelSelectorAsMap(broker.Spec.Selector)
	if err != nil {
		logrus.Errorf("LabelSelectorAsMap error,%v", err)
		return nil, err
	}
	err = c.List(ctx, list, client.MatchingLabels(selectorMap))
	if err != nil {
		logrus.Errorf("list storeset error,%v", err)
		return nil, err
	}
	if len(list.Items) <= 0 {
		return make([]*operator.StoreSet, 0), nil
	}

	return buildStoreData(list), nil
}

func buildStoreUri(item v13.StoreSet) []string {
	replicas := *item.Spec.Store.Replicas
	addrs := make([]string, replicas)
	var i int32
	for ; i < replicas; i++ {
		//TODO:重构名称的生成,应该和模板统一,使用template的自定义函数
		addrs[i] = fmt.Sprintf(`%s-%d.%s.%s:%s`, item.Name, i, item.Status.StoreStatus.ServiceName, item.Namespace, item.Spec.Store.Port)
	}
	return addrs
}

func buildStoreData(list *v13.StoreSetList) []*operator.StoreSet {
	var data = make([]*operator.StoreSet, len(list.Items))
	for _, item := range list.Items {
		data = append(data, &operator.StoreSet{
			Name:      item.Name,
			Namespace: item.Namespace,
			Uris:      buildStoreUri(item),
		})
	}
	return data
}

func getStatistics(pod v1.Pod) (*dispatcher.Statistics, error) {
	//发送请求
	url := fmt.Sprintf(statisticsUriFormat, pod.Status.PodIP, defaultPort)
	logrus.Infof("get statistics to %s", url)
	httpClient := &http.Client{Timeout: defaultTimeOut}
	resp, err := httpClient.Get(url)
	if err != nil {
		logrus.Errorf("Failed to get statistics from %s: %v", url, err)
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		logrus.Errorf("Failed to get statistics from %s statusCode: %v", url, resp.StatusCode)
		return nil, err
	}
	//parse body to Statistics
	statistics := &dispatcher.Statistics{}
	err = dispatcher.ParseStatistics(resp.Body, statistics)
	if err != nil {
		logrus.Errorf("Failed to parse statistics from %s: %v", url, err)
		return nil, err
	}
	return statistics, nil
}

func getMatchPod(ctx context.Context, c client.Client, request v12.Broker) (v1.Pod, error) {
	podList := &v1.PodList{}
	request.Labels["module"] = "dispatcher"
	request.Labels["broker"] = request.Name
	err := c.List(ctx, podList, client.InNamespace(request.Namespace), client.MatchingLabels(request.Labels))
	if err != nil {
		logrus.Errorf("Failed to list pods: %v", err)
		return v1.Pod{}, err
	}
	//filter pods with start time greater than current time minus 10 second
	add := time.Now().Add(-10 * time.Second)
	var pods []v1.Pod
	for _, pod := range podList.Items {
		if pod.Status.StartTime.After(add) {
			pods = append(pods, pod)
		}
	}
	//random pick one
	if len(pods) == 0 {
		logrus.Errorf("No pod found for broker %s", request.Name)
		return v1.Pod{}, err
	}
	//随机选择一个
	pod := pods[rand.Intn(len(pods))]
	return pod, nil
}
