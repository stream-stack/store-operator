# store-operator

## 启动k8s集群
```shell
kind create cluster --config deploy/kind-cluster.yaml --name c1
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

make docker-build docker-push IMG=ccr.ccs.tencentyun.com/stream/stream:latest
make deploy IMG=ccr.ccs.tencentyun.com/stream/stream:latest

kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/v1.7.1/cert-manager.yaml
```