package steps

import (
	"context"
	"fmt"
	"github.com/Jille/raftadmin/proto"
	v1 "github.com/stream-stack/store-operator/api/v1"
	"github.com/stream-stack/store-operator/controllers"
	"google.golang.org/grpc"
	v13 "k8s.io/api/apps/v1"
	v12 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"reflect"
	"strconv"
	"time"
)

func NewStoreSteps(cfg *controllers.InitConfig) {
	var service = &controllers.Step{
		Name: "storeService",
		GetObj: func() controllers.Object {
			return &v12.Service{}
		},
		Render: func(c *v1.StoreSet) controllers.Object {
			return &v12.Service{
				ObjectMeta: metav1.ObjectMeta{
					Name:        c.Name,
					Namespace:   c.Namespace,
					Labels:      c.Labels,
					Annotations: c.Annotations,
				},
				Spec: v12.ServiceSpec{
					Ports: []v12.ServicePort{
						{
							Name:       "grpc-store",
							Protocol:   v12.ProtocolTCP,
							Port:       containerPort.IntVal,
							TargetPort: containerPort,
						},
					},
					Selector: c.Labels,
					Type:     v12.ServiceTypeClusterIP,
				},
			}
		},
		SetStatus: func(c *v1.StoreSet, target, now controllers.Object) (needUpdate bool, updateObject controllers.Object, err error) {
			o := now.(*v12.Service)
			c.Status.StoreStatus.ServiceName = o.Name
			c.Status.StoreStatus.Service = o.Status

			t := target.(*v12.Service)
			if !reflect.DeepEqual(t.Spec, o.Spec) {
				o.Spec = t.Spec
				return true, o, nil
			}

			return false, now, nil
		},
		Next: func(c *v1.StoreSet) bool {
			return true
		},
		SetDefault: func(c *v1.StoreSet) {
			if c.Spec.Volume.LocalVolumeSource == nil {
				c.Spec.Volume.LocalVolumeSource = &v12.LocalVolumeSource{
					Path: DefaultVolumePath,
				}
			}
		},
	}
	var statefulsets = &controllers.Step{
		Name: "storeStatefulset",
		GetObj: func() controllers.Object {
			return &v13.StatefulSet{}
		},
		Render: func(c *v1.StoreSet) controllers.Object {
			filesystem := v12.PersistentVolumeFilesystem
			pvcName := c.Name
			volumeClaim := v12.PersistentVolumeClaim{
				ObjectMeta: metav1.ObjectMeta{
					Name:        pvcName,
					Namespace:   c.Namespace,
					Labels:      c.Labels,
					Annotations: c.Annotations,
				},
				Spec: v12.PersistentVolumeClaimSpec{
					AccessModes: []v12.PersistentVolumeAccessMode{v12.ReadWriteOnce},
					Selector: &metav1.LabelSelector{
						MatchLabels: c.Labels,
					},
					Resources: v12.ResourceRequirements{
						Requests: map[v12.ResourceName]resource.Quantity{
							v12.ResourceStorage: c.Spec.Volume.Capacity,
						},
					},
					//VolumeName:       "",
					StorageClassName: &c.Name,
					VolumeMode:       &filesystem,
				},
			}
			var terminationGracePeriodSeconds int64 = 30
			return &v13.StatefulSet{
				ObjectMeta: metav1.ObjectMeta{
					Name:        c.Name,
					Namespace:   c.Namespace,
					Labels:      c.Labels,
					Annotations: c.Annotations,
				},
				Spec: v13.StatefulSetSpec{
					Replicas: c.Spec.Store.Replicas,
					Selector: &metav1.LabelSelector{
						MatchLabels: c.Labels,
					},
					Template: v12.PodTemplateSpec{
						ObjectMeta: metav1.ObjectMeta{
							Labels:      c.Labels,
							Annotations: c.Annotations,
						},
						Spec: v12.PodSpec{
							Containers: []v12.Container{
								{
									Name:    "store",
									Image:   c.Spec.Store.Image,
									Command: []string{`store`},
									//      --Address string                  TCP host+port for this node (default "0.0.0.0:50051")
									//      --Bootstrap                       Whether to bootstrap the Raft cluster
									//      --DataDir string                  data dir (default "data")
									//      --RaftId string                   Node id used by Raft
									//      --Wal-BinaryLogFormat             wal BinaryLogFormat
									//      --Wal-NoSync                      wal NoSync
									//      --Wal-SegmentCacheSize int        wal SegmentCacheSize
									//      --Wal-SegmentSize int             wal SegmentSize (default 100000)
									//      --Grpc-ApplyLogTimeout duration   grpc apply log timeout second (default 1s)
									Args:       []string{`--DataDir=/data`, `--Bootstrap=true`},
									WorkingDir: "/",
									//TODO:集群内通信端口与 api端口分离
									Ports: []v12.ContainerPort{
										{
											Name:          "grpc-store",
											ContainerPort: containerPort.IntVal,
											Protocol:      v12.ProtocolTCP,
										},
									},
									//EnvFrom:    nil,
									//Env:        nil,
									//TODO:资源限制
									//Resources: v12.ResourceRequirements{},
									VolumeMounts: []v12.VolumeMount{
										{
											Name:      pvcName,
											ReadOnly:  false,
											MountPath: "/data",
										},
									},
									//TODO:健康检查升级TCP为GRPC(目前为了兼容旧版本k8s,所以不立即使用,Kubernetes v1.23 [alpha])
									//https://kubernetes.io/docs/tasks/configure-pod-container/configure-liveness-readiness-startup-probes/#define-a-grpc-liveness-probe
									LivenessProbe: &v12.Probe{
										ProbeHandler: v12.ProbeHandler{
											TCPSocket: &v12.TCPSocketAction{
												Port: containerPort,
											},
										},
										InitialDelaySeconds:           3,
										TimeoutSeconds:                5,
										PeriodSeconds:                 5,
										SuccessThreshold:              1,
										FailureThreshold:              3,
										TerminationGracePeriodSeconds: &terminationGracePeriodSeconds,
									},
									ReadinessProbe: &v12.Probe{
										ProbeHandler: v12.ProbeHandler{
											TCPSocket: &v12.TCPSocketAction{
												Port: containerPort,
											},
										},
										InitialDelaySeconds:           3,
										TimeoutSeconds:                5,
										PeriodSeconds:                 5,
										SuccessThreshold:              1,
										FailureThreshold:              3,
										TerminationGracePeriodSeconds: &terminationGracePeriodSeconds,
									},
									//StartupProbe:             nil,
									//Lifecycle:                nil,
									ImagePullPolicy: v12.PullIfNotPresent,
								},
							},
							RestartPolicy: v12.RestartPolicyOnFailure,
							//ServiceAccountName:            "",
							//ImagePullSecrets:              ,
							Affinity: &v12.Affinity{
								PodAntiAffinity: &v12.PodAntiAffinity{
									RequiredDuringSchedulingIgnoredDuringExecution: nil,
									PreferredDuringSchedulingIgnoredDuringExecution: []v12.WeightedPodAffinityTerm{
										{
											Weight: 100,
											PodAffinityTerm: v12.PodAffinityTerm{
												LabelSelector: &metav1.LabelSelector{
													MatchExpressions: []metav1.LabelSelectorRequirement{
														{
															Key:      AppLabelKey,
															Operator: metav1.LabelSelectorOpIn,
															Values:   []string{c.Labels[AppLabelKey]},
														},
													},
												},
												TopologyKey: topologyKeyHostname,
											},
										},
									},
								},
							},
							//SchedulerName:                 "",
							//Tolerations: nil,
							//ReadinessGates: nil,
						},
					},
					VolumeClaimTemplates: []v12.PersistentVolumeClaim{
						volumeClaim,
					},
					ServiceName: c.Name,
					UpdateStrategy: v13.StatefulSetUpdateStrategy{
						Type: v13.RollingUpdateStatefulSetStrategyType,
					},
					PersistentVolumeClaimRetentionPolicy: &v13.StatefulSetPersistentVolumeClaimRetentionPolicy{
						WhenDeleted: v13.DeletePersistentVolumeClaimRetentionPolicyType,
						WhenScaled:  v13.DeletePersistentVolumeClaimRetentionPolicyType,
					},
				},
			}
		},
		SetStatus: func(c *v1.StoreSet, target, now controllers.Object) (needUpdate bool, updateObject controllers.Object, err error) {
			o := now.(*v13.StatefulSet)
			c.Status.StoreStatus.Workload = o.Status
			c.Status.StoreStatus.WorkloadName = o.Name

			t := target.(*v13.StatefulSet)
			if !reflect.DeepEqual(t.Spec, o.Spec) {
				o.Spec = t.Spec
				return true, o, nil
			}

			return false, now, nil
		},
		Next: func(c *v1.StoreSet) bool {
			replicas := *c.Spec.Store.Replicas
			if c.Status.StoreStatus.Workload.AvailableReplicas != replicas {
				return false
			}
			infos := getLeaderInfos(c, replicas)
			if len(infos) == 1 {
				return true
			}
			leader := getMaxTermLeader(infos)
			if leader == nil {
				//没有leader
				return false
			}
			err := joinLeader(leader, infos)
			if err != nil {
				return false
			}
			closeConn(infos)
			return true
		},
		SetDefault: func(c *v1.StoreSet) {
			if c.Spec.Store.Image == "" {
				c.Spec.Store.Image = cfg.StoreImage
			}
			if c.Spec.Store.Replicas == nil {
				c.Spec.Store.Replicas = &cfg.StoreReplicas
			}

		},
		ValidateCreateStep: func(c *v1.StoreSet) field.ErrorList {
			var allErrs field.ErrorList
			if *c.Spec.Store.Replicas%2 != 0 {
				allErrs = append(allErrs, field.Invalid(field.NewPath("spec.store.replicas"), c.Spec.Store.Replicas, "必须为奇数"))
			}
			return allErrs
		},
		ValidateUpdateStep: func(now *v1.StoreSet, old *v1.StoreSet) field.ErrorList {
			var allErrs field.ErrorList
			if *now.Spec.Store.Replicas%2 != 0 {
				allErrs = append(allErrs, field.Invalid(field.NewPath("spec.store.replicas"), now.Spec.Store.Replicas, "必须为奇数"))
			}
			return allErrs
		},
	}

	controllers.AddSteps(cfg.Version, &controllers.Step{
		Order: 20,
		Name:  "store",
		Sub:   []*controllers.Step{service, statefulsets},
	})
}

func closeConn(infos []*leaderInfo) {
	for _, info := range infos {
		_ = info.conn.Close()
	}
}

func joinLeader(leader *leaderInfo, infos []*leaderInfo) error {
	for _, info := range infos {
		if info.addr == leader.addr {
			continue
		}
		_, err := leader.client.AddVoter(context.TODO(), &proto.AddVoterRequest{
			Id:            info.hostname,
			Address:       info.addr,
			PreviousIndex: 0,
		})
		if err != nil {
			return err
		}
	}
	return nil
}

func getMaxTermLeader(infos []*leaderInfo) *leaderInfo {
	if len(infos) == 0 {
		return nil
	}
	var result = infos[0]
	for _, info := range infos {
		if result.term < info.term {
			result = info
		}
	}
	return result
}

func getLeaderInfos(c *v1.StoreSet, replicas int32) []*leaderInfo {
	var i int32
	leaders := make([]*leaderInfo, 0)
	for ; i < replicas; i++ {
		hostname := fmt.Sprintf(`%s-%d`, c.Status.StoreStatus.WorkloadName, i)
		address := fmt.Sprintf(`%s.%s.%s.svc:%d`, hostname, c.Status.StoreStatus.ServiceName, c.Namespace, containerPort.IntVal)

		timeout, _ := context.WithTimeout(context.TODO(), defaultGrpcTimeOut)
		conn, err := grpc.DialContext(timeout, address, grpc.WithInsecure())
		if err != nil {
			continue
		}

		client := proto.NewRaftAdminClient(conn)
		stats, err := client.Stats(timeout, &proto.StatsRequest{})
		if err != nil {
			_ = conn.Close()
			continue
		}
		if stats.Stats[`state`] == `Leader` {
			atoi, _ := strconv.Atoi(stats.Stats[`term`])
			leaders = append(leaders, &leaderInfo{
				hostname: hostname,
				addr:     address,
				conn:     conn,
				term:     atoi,
				client:   client,
			})
		} else {
			_ = conn.Close()
		}
	}
	return leaders
}

const topologyKeyHostname = "kubernetes.io/hostname"
const defaultGrpcTimeOut = time.Second * 5

var containerPort = intstr.FromInt(50001)

type leaderInfo struct {
	addr     string
	conn     *grpc.ClientConn
	term     int
	client   proto.RaftAdminClient
	hostname string
}
