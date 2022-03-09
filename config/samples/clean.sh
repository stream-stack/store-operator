cd ../../ && make undeploy

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

docker exec -it c1-worker /bin/bash -c 'cd /data/ && rm -rf *'
docker exec -it c1-worker2 /bin/bash -c 'cd /data/ && rm -rf *'
docker exec -it c1-worker3 /bin/bash -c 'cd /data/ && rm -rf *'


docker exec  c1-worker /bin/bash -c 'cd /data/wal && cat * | wc -l'
docker exec  c1-worker2 /bin/bash -c 'cd /data/wal && cat * | wc -l'
docker exec  c1-worker3 /bin/bash -c 'cd /data/wal && cat * | wc -l'