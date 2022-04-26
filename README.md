# store-operator

## Overview

stream stack 是一个用于提供流式数据存储的服务
目前还在开发阶段

## 安装

### 安装k8s集群(如果没有)

```shell
wget -O kind-cluster.yaml https://raw.githubusercontent.com/stream-stack/store-operator/main/config/samples/kind-cluster.yaml
kind create cluster --config kind-cluster.yaml --name c1
```

该命令使用kind创建了1个master,3个node节点的k8s集群
因storeset会使用localpersistent volume来存储数据,所以需要在k8s集群中创建localpersistent volume 文件夹

```shell
docker exec -it c1-worker mkdir -p /data
docker exec -it c1-worker2 mkdir -p /data
docker exec -it c1-worker3 mkdir -p /data
```

### 安装cert-manager

```shell
kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/v1.7.1/cert-manager.yaml
```

### 安装operator

```shell
kubectl apply -f https://raw.githubusercontent.com/stream-stack/store-operator/main/deploy/operator-install.yaml
```

## 使用

### 创建存储事件的storeset

```shell
kubectl apply -f https://raw.githubusercontent.com/stream-stack/store-operator/main/config/samples/core_v1_storeset.yaml
```

该命令创建了一个storeset,该storeset的名称为test,label为test,后续的broker和subscription都会使用该storeset

### 创建broker

```shell
kubectl apply -f https://raw.githubusercontent.com/stream-stack/store-operator/main/config/samples/knative_v1_broker.yaml

```

该命令创建了一个broker,该broker的名称为test,label为test,后续的subscription都会使用该broker

### 创建事件接收容器

```shell
kubectl apply -f https://raw.githubusercontent.com/stream-stack/store-operator/main/config/samples/display.yaml
```

该命令创建了一个事件接收容器,用来接收标准的cloudevents事件,并在日志中显示
可通过`kubectl logs -f hello-display`查看日志

### 创建subscription

```shell
kubectl apply -f https://raw.githubusercontent.com/stream-stack/store-operator/main/config/samples/knative_v1_subscription.yaml
```

### 发送事件

#### 获取broker地址

```shell
kubectl  get svc test-dispatcher
###样例输出
NAME              TYPE        CLUSTER-IP      EXTERNAL-IP   PORT(S)   AGE
test-dispatcher   ClusterIP   10.96.150.185   <none>        80/TCP    38m
```

其中10.96.150.185为broker的地址,可以通过该地址发送事件

```shell
curl -X POST -v --location "http://10.96.150.185" \
    -H "Content-Type: application/json" \
    -H "Ce-Id: 1" \
    -H "Ce-Specversion: 1.0" \
    -H "Ce-Type: test" \
    -H "Ce-Source: testSource" \
    -H "Content-Type: application/json" \
    -d "{
          \"test\":1
        }"

curl -X POST -v --location "http://10.96.150.185" \
    -H "Content-Type: application/json" \
    -H "Ce-Id: 2" \
    -H "Ce-Specversion: 1.0" \
    -H "Ce-Type: test" \
    -H "Ce-Source: testSource" \
    -H "Content-Type: application/json" \
    -d "{
          \"test\":2
        }"

```

通过kubectl logs -f hello-display 查看日志

## 卸载

```shell
kubectl delete -f https://raw.githubusercontent.com/stream-stack/store-operator/main/config/samples/display.yaml
kubectl delete -f https://raw.githubusercontent.com/stream-stack/store-operator/main/config/samples/knative_v1_subscription.yaml
kubectl delete -f https://raw.githubusercontent.com/stream-stack/store-operator/main/config/samples/knative_v1_broker.yaml
kubectl delete -f https://raw.githubusercontent.com/stream-stack/store-operator/main/config/samples/core_v1_storeset.yaml
kubectl delete -f https://raw.githubusercontent.com/stream-stack/store-operator/main/deploy/operator-install.yaml
kubectl delete -f https://github.com/cert-manager/cert-manager/releases/download/v1.7.1/cert-manager.yaml

kubectl patch pvc test-test-0 -p '{"metadata":{"finalizers":null}}'
kubectl patch pvc test-test-1 -p '{"metadata":{"finalizers":null}}'
kubectl patch pvc test-test-2 -p '{"metadata":{"finalizers":null}}'
kubectl delete pvc test-test-0 test-test-1 test-test-2

kubectl patch pv test-0 -p '{"metadata":{"finalizers":null}}'
kubectl patch pv test-1 -p '{"metadata":{"finalizers":null}}'
kubectl patch pv test-2 -p '{"metadata":{"finalizers":null}}'
kubectl delete pv test-0 test-1 test-2
```

## 启动k8s集群

```shell
kind create cluster --config config/samples/kind-cluster.yaml --name c1

kubebuilder init --domain my.domain --repo my.domain/guestbook
kubebuilder init --domain stream-stack.tanx --repo github.com/stream-stack/store-operator

kubebuilder create api --group core --version v1 --kind StoreSet
kubebuilder create api --group knative --version v1 --kind Broker
kubebuilder create api --group knative --version v1 --kind Subscription

kubebuilder create webhook --group core --version v1 --kind StoreSet --defaulting --programmatic-validation
kubebuilder create webhook --group knative --version v1 --kind Broker --defaulting --programmatic-validation
kubebuilder create webhook --group knative --version v1 --kind Subscription --defaulting --programmatic-validation

```

```shell
make manifests
make install

make run ENABLE_WEBHOOKS=false


kubectl delete storeset test
kubectl patch pvc test-test-0 -p '{"metadata":{"finalizers":null}}'
kubectl patch pvc test-test-1 -p '{"metadata":{"finalizers":null}}'
kubectl patch pvc test-test-2 -p '{"metadata":{"finalizers":null}}'
kubectl delete pvc test-test-0 test-test-1 test-test-2

kubectl patch pv test-0 -p '{"metadata":{"finalizers":null}}'
kubectl patch pv test-1 -p '{"metadata":{"finalizers":null}}'
kubectl patch pv test-2 -p '{"metadata":{"finalizers":null}}'
kubectl delete pv test-0 test-1 test-2

kubectl get pod,sts,pv,pvc,svc
kubectl delete storeset test

kubectl apply -f core_v1_storeset.yaml

make docker-build docker-push IMG=ccr.ccs.tencentyun.com/stream/operator:latest
make deploy IMG=ccr.ccs.tencentyun.com/stream/operator:latest

kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/v1.7.1/cert-manager.yaml
```