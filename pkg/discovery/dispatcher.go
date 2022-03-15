package discovery

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	_ "github.com/Jille/grpc-multi-resolver"
	grpc_retry "github.com/grpc-ecosystem/go-grpc-middleware/retry"
	v1 "github.com/stream-stack/store-operator/apis/knative/v1"
	v15 "github.com/stream-stack/store-operator/apis/storeset/v1"
	"github.com/stream-stack/store-operator/pkg/base"
	protocol "github.com/stream-stack/store-operator/pkg/proto"
	"google.golang.org/grpc"
	_ "google.golang.org/grpc/health"
	"io/ioutil"
	v13 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/rand"
	"net/http"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"strings"
	"time"
)

func StartDispatcherStoreSetDiscovery(ctx *base.StepContext, broker *v1.Broker) error {
	var err error
	list := &v15.StoreSetList{}
	selectorMap, err := v13.LabelSelectorAsMap(broker.Spec.Selector)
	if err != nil {
		return err
	}
	err = ctx.GetClient().List(ctx, list, client.MatchingLabels(selectorMap))
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
	var result []byte
	for _, add := range adds {
		//推送store到dispatcher
		r, err := sendRequest(add, buffer)
		if err != nil {
			return err
		}
		result = r
	}

	configuration := &Configuration{}
	err = json.Unmarshal(result, configuration)
	if err != nil {
		return err
	}
	if len(configuration.Partitions) > 0 {
		return nil
	}
	//分配第一个分片并推送分片
	return allocatePartition(ctx, list.Items, storesetData, broker)
}

func allocatePartition(ctx *base.StepContext, items []v15.StoreSet, data []store, broker *v1.Broker) error {
	//TODO:根据分片规则分片
	intn := rand.Intn(len(items))
	//set := items[intn]
	s := data[intn]
	buffer := &bytes.Buffer{}
	err := binary.Write(buffer, binary.BigEndian, Partition{
		Begin: "0",
		Store: s,
	})
	if err != nil {
		return err
	}

	serviceConfig := `{"healthCheckConfig": {"serviceName": "store"}, "loadBalancingConfig": [ { "round_robin": {} } ]}`
	retryOpts := []grpc_retry.CallOption{
		grpc_retry.WithBackoff(grpc_retry.BackoffExponential(100 * time.Millisecond)),
		grpc_retry.WithMax(5),
	}
	conn, err := grpc.Dial("dns:///"+strings.Join(s.Uris, ","),
		grpc.WithDefaultServiceConfig(serviceConfig), grpc.WithInsecure(),
		grpc.WithDefaultCallOptions(grpc.WaitForReady(true)),
		grpc.WithUnaryInterceptor(grpc_retry.UnaryClientInterceptor(retryOpts...)))
	if err != nil {
		return err
	}
	defer conn.Close()
	eventCli := protocol.NewEventServiceClient(conn)
	apply, err := eventCli.Apply(ctx.Context, &protocol.ApplyRequest{
		StreamName: "_system",
		StreamId:   fmt.Sprintf("%s-%s", broker.Namespace, broker.Name),
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

func sendRequest(add string, buffer *bytes.Buffer) ([]byte, error) {
	c := http.Client{Timeout: time.Second * 5}
	post, err := c.Post(add, `application/json`, buffer)
	if err != nil {
		return nil, err
	}
	defer post.Body.Close()
	if post.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("http status not ok(200),current:%v", post.StatusCode)
	}
	all, err := ioutil.ReadAll(post.Body)
	if err != nil {
		return nil, err
	}
	fmt.Println("推送store后返回值: ", string(all))
	return all, nil
}

type store struct {
	Name      string   `json:"name"`
	Namespace string   `json:"namespace"`
	Uris      []string `json:"uris"`
}

type Partition struct {
	Begin string `json:"begin"`
	Store store  `json:"store"`
}

type Configuration struct {
	Stores     []store     `json:"stores"`
	Partitions []Partition `json:"partitions"`
	MaxEventId string      `json:"max_event_id"`
}

func buildStoreData(list *v15.StoreSetList) []store {
	data := make([]store, len(list.Items))
	for i, item := range list.Items {
		data[i] = store{Uris: buildStoreUri(item), Name: item.Name, Namespace: item.Namespace}
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
		//TODO:重构名称的生成,应该和模板统一,使用template的自定义函数
		addrs[i] = fmt.Sprintf(`%s-dispatcher-%d.%s.%s`, broker.Name, i, broker.Status.Dispatcher.SvcName, broker.GetNamespace())
	}
	return addrs
}
