package consts

const (
	// NodePoolNameLabel uses to label ConfigMap objects which are associated with the NodePool
	NodePoolNameLabel = "hypershift.openshift.io/nodePool"

	// PerformanceProfileNameLabel uses to label a ConfigMaps that hold objects which were
	// created by the performanceProfile controller and which are associated with the performance-profile
	// name mentioned in the label
	PerformanceProfileNameLabel = "hypershift.openshift.io/performanceProfileName"

	// ControllerGeneratedTunedConfigMapLabel uses to label a ConfigMap that holds encoded tuned object
	ControllerGeneratedTunedConfigMapLabel = "hypershift.openshift.io/tuned-config"

	// ControllerGeneratedPerformanceProfileConfigMapLabel uses
	//to label a ConfigMap that holds encoded performance-profile object
	ControllerGeneratedPerformanceProfileConfigMapLabel = "hypershift.openshift.io/performanceprofile-config"

	// NTOGeneratedMachineConfigLabel attached to kubelet and machine config configmaps generated by NTO.
	// The Hypershift operator watches these and acts upon their creation/update/deletion.
	NTOGeneratedMachineConfigLabel = "hypershift.openshift.io/nto-generated-machine-config"

	// NTOGeneratedPerformanceProfileStatusConfigMapLabel uses
	//to label a ConfigMap that holds encoded performance-profile status object
	NTOGeneratedPerformanceProfileStatusConfigMapLabel = "hypershift.openshift.io/nto-generated-performance-profile-status"

	// KubeletConfigConfigMapLabel uses
	// to label a ConfigMap that holds a KubeletConfig object
	KubeletConfigConfigMapLabel = "hypershift.openshift.io/kubeletconfig-config"

	// TuningKey is the key under ConfigMap.Data on which encoded
	// tuned and performance-profile objects are stored.
	TuningKey = "tuning"

	// ConfigKey is the key under ConfigMap.Data on which encoded
	// machine-config, kubelet-config and container-runtime-config objects are stored.
	ConfigKey = "config"

	// PerformanceProfileStatusKey is the key under ConfigMap.Data on which an encoded
	// performance-profile status object is stored.
	PerformanceProfileStatusKey = "status"
)
