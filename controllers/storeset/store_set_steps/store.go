package store_set_steps

import (
	"context"
	"fmt"
	"github.com/Jille/raftadmin/proto"
	"github.com/go-logr/logr"
	configv1 "github.com/stream-stack/store-operator/apis/config/v1"
	v14 "github.com/stream-stack/store-operator/apis/storeset/v1"
	"github.com/stream-stack/store-operator/pkg/base"
	"google.golang.org/grpc"
	v13 "k8s.io/api/apps/v1"
	v12 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"reflect"
	"strconv"
)

func NewStoreSteps(cfg configv1.StreamControllerConfig) *base.Step {
	var service = &base.Step{
		Name: "storeService",
		GetObj: func() base.StepObject {
			return &v12.Service{}
		},
		Render: func(set base.StepObject) (base.StepObject, error) {
			c := set.(*v14.StoreSet)
			storeContainerPort := intstr.Parse(c.Spec.Store.Port)
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
							Port:       storeContainerPort.IntVal,
							TargetPort: storeContainerPort,
						},
					},
					Selector: c.Labels,
					Type:     v12.ServiceTypeClusterIP,
				},
			}, nil
		},
		SetStatus: func(set base.StepObject, target, now base.StepObject) (needUpdate bool, updateObject base.StepObject, err error) {
			c := set.(*v14.StoreSet)
			o := now.(*v12.Service)
			c.Status.StoreStatus.ServiceName = o.Name
			c.Status.StoreStatus.Service = o.Status

			t := target.(*v12.Service)
			if !reflect.DeepEqual(t.Spec.Selector, o.Spec.Selector) {
				o.Spec.Selector = t.Spec.Selector
				return true, o, nil
			}
			if !reflect.DeepEqual(t.Spec.Ports, o.Spec.Ports) {
				o.Spec.Ports = t.Spec.Ports
				return true, o, nil
			}
			if !reflect.DeepEqual(t.Spec.Type, o.Spec.Type) {
				o.Spec.Type = t.Spec.Type
				return true, o, nil
			}
			c.Status.Status = v14.StoreSetStatusStoreSetSvcCreating
			return false, now, nil
		},
		Next: func(ctx *base.StepContext) (bool, error) {
			return true, nil
		},
		SetDefault: func(set base.StepObject) {
			c := set.(*v14.StoreSet)
			if c.Spec.Volume.LocalVolumeSource == nil {
				c.Spec.Volume.LocalVolumeSource = &v12.LocalVolumeSource{
					Path: cfg.Store.VolumePath,
				}
			}
		},
	}
	var statefulsets = &base.Step{
		Name: "storeStatefulset",
		GetObj: func() base.StepObject {
			return &v13.StatefulSet{}
		},
		Render: func(set base.StepObject) (base.StepObject, error) {
			c := set.(*v14.StoreSet)
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
			storeContainerPort := intstr.Parse(c.Spec.Store.Port)
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
									Command: []string{`/store`},
									//--Address string                       TCP host+port for this node (default "0.0.0.0:50051")
									//--BaseDir string                       base data directory (default "./data")
									//--Bootstrap                            Whether to bootstrap the Raft cluster
									//--DataDir string                       data dir (default "data")
									//--Grpc-ApplyLogTimeout duration        grpc apply log timeout second (default 1s)
									//--RaftId string                        Node id used by Raft(default hostname)
									//--ValueDir string                      value dir (default "value")
									//--badgerGCInterval duration            badger GCInterval (default 1m0s)
									//--badgerGCThreshold int                badger GCThreshold (default 1073741824)
									//--badgerMandatoryGCInterval duration   badger MandatoryGCInterval (default 10m0s)
									//--badgerSync                           badger Sync (default true)
									//--badgerValueLogGC                     badger ValueLogGC
									//-h, --help                                 help for this command
									//--snapshotDataDir string               snapshot data directory (default "snapshot")
									//--snapshotRetain int                   snapshot retain count (default 10)
									Args:       []string{`--BaseDir=/data`, `--Bootstrap=true`},
									WorkingDir: "/",
									//TODO:???????????????????????? api????????????
									Ports: []v12.ContainerPort{
										{
											Name:          "grpc-store",
											ContainerPort: storeContainerPort.IntVal,
											Protocol:      v12.ProtocolTCP,
										},
									},
									//EnvFrom:    nil,
									//Env:        nil,
									//TODO:????????????
									//Resources: v12.ResourceRequirements{},
									VolumeMounts: []v12.VolumeMount{
										{
											Name:      pvcName,
											ReadOnly:  false,
											MountPath: "/data",
										},
									},
									//TODO:??????????????????TCP???GRPC(???????????????????????????k8s,?????????????????????,Kubernetes v1.23 [alpha])
									//https://kubernetes.io/docs/tasks/configure-pod-container/configure-liveness-readiness-startup-probes/#define-a-grpc-liveness-probe
									LivenessProbe: &v12.Probe{
										ProbeHandler: v12.ProbeHandler{
											TCPSocket: &v12.TCPSocketAction{
												Port: storeContainerPort,
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
												Port: storeContainerPort,
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
							RestartPolicy: v12.RestartPolicyAlways,
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
												TopologyKey: cfg.Store.TopologyKey,
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
			}, nil
		},
		SetStatus: func(set base.StepObject, target, now base.StepObject) (needUpdate bool, updateObject base.StepObject, err error) {
			c := set.(*v14.StoreSet)
			o := now.(*v13.StatefulSet)
			c.Status.StoreStatus.Workload = o.Status
			c.Status.StoreStatus.WorkloadName = o.Name

			t := target.(*v13.StatefulSet)
			if !reflect.DeepEqual(t.Spec, o.Spec) {
				o.Spec = t.Spec
				return true, o, nil
			}
			c.Status.Status = v14.StoreSetStatusStoreSetStsCreating

			return false, now, nil
		},
		Next: func(ctx *base.StepContext) (bool, error) {
			c := ctx.StepObject.(*v14.StoreSet)
			replicas := *c.Spec.Store.Replicas

			if c.Status.StoreStatus.Workload.ReadyReplicas != replicas {
				return false, nil
			}
			infos, err := getNodeInfos(ctx.Logger, c, replicas, cfg.Store)
			if err != nil {
				return false, err
			}
			defer closeConn(infos)
			ctx.Logger.Info("??????????????????", ctx.StepObject.GetName(), len(infos))
			leader := getMaxAppliedIndexLeader(infos)
			//??????leader
			if leader == nil {
				ctx.Logger.Info("??????leader?????????0,?????????pod????????????")
				return false, fmt.Errorf("??????leader?????????0,?????????pod????????????")
			}
			ctx.Logger.Info("?????????leader,????????????????????????????????????leader", "address", infos[0].addr)
			err = joinLeader(ctx.Logger, leader, infos, cfg.Store.GrpcTimeOut)
			if err != nil {
				return false, err
			}
			c.Status.Status = v14.StoreSetStatusReady

			return true, nil
		},
		SetDefault: func(set base.StepObject) {
			c := set.(*v14.StoreSet)
			if c.Spec.Store.Image == "" {
				c.Spec.Store.Image = cfg.Store.Image
			}
			if c.Spec.Store.Replicas == nil {
				c.Spec.Store.Replicas = &cfg.Store.Replicas
			}
			if len(c.Spec.Store.Port) == 0 {
				c.Spec.Store.Port = cfg.Store.Port
			}

		},
		ValidateCreateStep: func(set base.StepObject) field.ErrorList {
			c := set.(*v14.StoreSet)
			var allErrs field.ErrorList
			if *c.Spec.Store.Replicas%2 == 0 {
				allErrs = append(allErrs, field.Invalid(field.NewPath("spec.store.replicas"), c.Spec.Store.Replicas, "???????????????"))
			}
			return allErrs
		},
		ValidateUpdateStep: func(nowSet base.StepObject, oldSet base.StepObject) field.ErrorList {
			now := nowSet.(*v14.StoreSet)
			old := oldSet.(*v14.StoreSet)
			var allErrs field.ErrorList
			if *now.Spec.Store.Replicas%2 == 0 {
				allErrs = append(allErrs, field.Invalid(field.NewPath("spec.store.replicas"), now.Spec.Store.Replicas, "???????????????"))
			}
			if *now.Spec.Store.Replicas < *old.Spec.Store.Replicas {
				allErrs = append(allErrs, field.Invalid(field.NewPath("spec.store.replicas"), now.Spec.Store.Replicas, "???????????????"))
			}
			return allErrs
		},
	}

	return &base.Step{
		Name: "store",
		Sub:  []*base.Step{service, statefulsets},
	}
}

func closeConn(infos []*nodeInfo) {
	for _, info := range infos {
		if info.conn != nil {
			_ = info.conn.Close()
		}
	}
}

func joinLeader(logger logr.Logger, leader *nodeInfo, infos []*nodeInfo, t metav1.Duration) error {
	timeout, cancelFunc := context.WithTimeout(context.TODO(), t.Duration)
	defer cancelFunc()
	configuration, err := leader.client.GetConfiguration(timeout, &proto.GetConfigurationRequest{})
	if err != nil {
		logger.Info("Leader Grpc Call GetConfiguration Error", "error", err)
		return err
	}
	for _, info := range infos {
		if info.addr == leader.addr {
			logger.Info("target address is leader address , ignore", "leader", leader.addr, "target", info.addr)
			continue
		}
		exist := false
		for _, server := range configuration.Servers {
			if server.Address == info.addr {
				exist = true
				break
			}
		}
		if exist {
			logger.Info("target address exist leader configuration, ignore")
			continue
		}
		_, err = leader.client.AddVoter(timeout, &proto.AddVoterRequest{
			Id:      info.hostname,
			Address: info.addr,
		})
		if err != nil {
			logger.Info("leader join node error", "error", err, "target", info.addr)
			return err
		}
		logger.Info("leader join node success", "target", info.addr)
	}
	return nil
}

func getMaxAppliedIndexLeader(infos []*nodeInfo) *nodeInfo {
	if len(infos) == 0 {
		return nil
	}
	var result = infos[0]
	for _, info := range infos {
		if result.appliedIndex < info.appliedIndex && info.state == `Leader` {
			result = info
		}
	}
	return result
}

func getNodeInfos(logger logr.Logger, c *v14.StoreSet, replicas int32, cfg configv1.StoreDefault) ([]*nodeInfo, error) {
	var i int32
	nodes := make([]*nodeInfo, 0)
	for ; i < replicas; i++ {
		hostname := fmt.Sprintf(`%s-%d`, c.Status.StoreStatus.WorkloadName, i)
		address := fmt.Sprintf(`%s.%s.%s.svc:%s`, hostname, c.Status.StoreStatus.ServiceName, c.Namespace, cfg.Port)

		logger.Info("try connect statefulset pod", "address", address)
		timeout, cancelFunc := context.WithTimeout(context.TODO(), cfg.GrpcTimeOut.Duration)
		defer cancelFunc()
		conn, err := grpc.DialContext(timeout, address, grpc.WithInsecure())
		if err != nil {
			logger.Info("Grpc Dial error", "error", err, "address", address)
			return nodes, err
		}

		client := proto.NewRaftAdminClient(conn)
		stats, err := client.Stats(timeout, &proto.StatsRequest{})
		if err != nil {
			logger.Info("Grpc call Status method error", "error", err, "address", address)
			_ = conn.Close()
			return nodes, err
		}
		atoi, _ := strconv.Atoi(stats.Stats[`applied_index`])
		nodes = append(nodes, &nodeInfo{
			hostname:     hostname,
			addr:         address,
			conn:         conn,
			appliedIndex: atoi,
			state:        stats.Stats[`state`],
			client:       client,
		})
	}
	return nodes, nil
}

type nodeInfo struct {
	addr         string
	conn         *grpc.ClientConn
	appliedIndex int
	client       proto.RaftAdminClient
	hostname     string
	state        string
}
