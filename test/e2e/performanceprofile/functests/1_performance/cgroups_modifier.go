package __performance

import (
	"context"
	"fmt"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	testclient "github.com/openshift/cluster-node-tuning-operator/test/e2e/performanceprofile/functests/utils/client"
	"github.com/openshift/cluster-node-tuning-operator/test/e2e/performanceprofile/functests/utils/namespaces"
	"github.com/openshift/cluster-node-tuning-operator/test/e2e/performanceprofile/functests/utils/pods"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"k8s.io/kubernetes/pkg/kubelet/cm/cpuset"
	"strings"
	"time"
)

const (
	sharedPodName = "shared-pod"
	sharedCntName = "shared-cnt"
)

var _ = Describe("[performance] cgroups modifier", func() {
	BeforeEach(func() {
		pod := pods.GetTestPod(
			pods.WithResources(
				corev1.ResourceRequirements{
					Limits: corev1.ResourceList{
						corev1.ResourceCPU:    resource.MustParse("2"),
						corev1.ResourceMemory: resource.MustParse("100Mi"),
					},
				}),
			pods.WithContainerName(sharedCntName))
		pod.Name = sharedPodName

		err := testclient.Client.Create(context.TODO(), pod)
		Expect(err).ToNot(HaveOccurred())
		Expect(pods.WaitForPhase(pod, corev1.PodRunning, time.Minute)).ToNot(HaveOccurred())
	})

	AfterEach(func() {
		pod := pods.GetTestPod()
		pod.Name = sharedPodName
		err := testclient.Client.Delete(context.TODO(), pod)
		Expect(err).ToNot(HaveOccurred())
		Expect(pods.WaitForDeletion(pod, time.Minute)).ToNot(HaveOccurred())
	})

	When("a pod gets created with the shared annotations", func() {
		It("should contains the shared cpus under its cgroups", func() {
			By("creating a pod with cpu-shared-container annotation")
			pod := pods.GetTestPod(
				pods.WithAnnotations(map[string]string{"cpu-shared-container.crio.io": sharedCntName}),
				pods.WithResources(corev1.ResourceRequirements{
					Limits: corev1.ResourceList{
						corev1.ResourceCPU:    resource.MustParse("2"),
						corev1.ResourceMemory: resource.MustParse("100Mi"),
					},
				}))

			err := testclient.Client.Create(context.TODO(), pod)
			Expect(err).ToNot(HaveOccurred())
			Expect(pods.WaitForPhase(pod, corev1.PodRunning, time.Minute)).ToNot(HaveOccurred())

			//cmd := exec.Cmd{
			//	Path: "/usr/bin/sleep",
			//	Args: []string{"sleep", "300"},
			//}
			//Expect(cmd.Run()).ToNot(HaveOccurred())

			key := types.NamespacedName{
				Namespace: namespaces.TestingNamespace.Name,
				Name:      sharedPodName,
			}
			sharedPod := &corev1.Pod{}
			err = testclient.Client.Get(context.TODO(), key, sharedPod)
			Expect(err).ToNot(HaveOccurred())

			By("inspecting the pod's accessible cpus")
			cpus, err := getCpusFrom(testclient.K8sClient, pod)
			Expect(err).ToNot(HaveOccurred())

			sharedCpus, err := getCpusFrom(testclient.K8sClient, sharedPod)
			Expect(err).ToNot(HaveOccurred())
			intersect := cpus.Intersection(*sharedCpus)
			Expect(intersect.Equals(*sharedCpus)).To(BeTrue(), "shared cpus: %s, are not present under pod: %v", sharedCpus.String(), fmt.Sprintf("%s/%s", pod.Namespace, pod.Name))

			By(fmt.Sprintf("deleting the test pod: %s", pod.Name))
			err = testclient.Client.Delete(context.TODO(), pod)
			Expect(err).ToNot(HaveOccurred())
			Expect(pods.WaitForDeletion(pod, time.Minute)).ToNot(HaveOccurred())
		})

		// this test intended to make sure new pods are not triggering
		// the CPU manager conciliation loop to remove the shared CPUs
		// from the pod's cgroups
		It("should still access the shared cpus when new pods gets created", func() {
			By("creating a pod with cpu-shared-container annotation")
			pod := pods.GetTestPod(
				pods.WithAnnotations(map[string]string{"cpu-shared-container.crio.io": sharedCntName}),
				pods.WithResources(corev1.ResourceRequirements{
					Limits: corev1.ResourceList{
						corev1.ResourceCPU:    resource.MustParse("2"),
						corev1.ResourceMemory: resource.MustParse("100Mi"),
					},
				}))

			err := testclient.Client.Create(context.TODO(), pod)
			Expect(err).ToNot(HaveOccurred())
			Expect(pods.WaitForPhase(pod, corev1.PodRunning, time.Minute)).ToNot(HaveOccurred())

			pod2 := pods.GetTestPod()
			err = testclient.Client.Create(context.TODO(), pod2)
			Expect(err).ToNot(HaveOccurred())
			Expect(pods.WaitForPhase(pod2, corev1.PodRunning, time.Minute)).ToNot(HaveOccurred())

			key := types.NamespacedName{
				Namespace: namespaces.TestingNamespace.Name,
				Name:      sharedPodName,
			}
			sharedPod := &corev1.Pod{}
			err = testclient.Client.Get(context.TODO(), key, sharedPod)
			Expect(err).ToNot(HaveOccurred())

			By("inspecting the pod's accessible cpus")
			cpus, err := getCpusFrom(testclient.K8sClient, pod)
			Expect(err).ToNot(HaveOccurred())

			sharedCpus, err := getCpusFrom(testclient.K8sClient, sharedPod)
			Expect(err).ToNot(HaveOccurred())
			intersect := cpus.Intersection(*sharedCpus)
			Expect(intersect.Equals(*sharedCpus)).To(BeTrue(), "shared cpus: %s, are not present under pod: %v", sharedCpus.String(), fmt.Sprintf("%s/%s", pod.Namespace, pod.Name))

			By(fmt.Sprintf("deleting the test pod: %s", pod.Name))
			err = testclient.Client.Delete(context.TODO(), pod)
			Expect(err).ToNot(HaveOccurred())
			Expect(pods.WaitForDeletion(pod, time.Minute)).ToNot(HaveOccurred())

			By(fmt.Sprintf("deleting the second test pod: %s", pod2.Name))
			err = testclient.Client.Delete(context.TODO(), pod2)
			Expect(err).ToNot(HaveOccurred())
			Expect(pods.WaitForDeletion(pod2, time.Minute)).ToNot(HaveOccurred())
		})

		It("should still access the shared cpus when new pod try asking for it as well", func() {
			pod := pods.GetTestPod(
				pods.WithAnnotations(map[string]string{"cpu-shared-container.crio.io": sharedCntName}),
				pods.WithResources(corev1.ResourceRequirements{
					Limits: corev1.ResourceList{
						corev1.ResourceCPU:    resource.MustParse("2"),
						corev1.ResourceMemory: resource.MustParse("100Mi"),
					},
				}))

			pod2 := pod.DeepCopy()
			err := testclient.Client.Create(context.TODO(), pod)
			Expect(err).ToNot(HaveOccurred())
			Expect(pods.WaitForPhase(pod, corev1.PodRunning, time.Minute)).ToNot(HaveOccurred())

			err = testclient.Client.Create(context.TODO(), pod2)
			Expect(err).ToNot(HaveOccurred())
			Expect(pods.WaitForPhase(pod2, corev1.PodRunning, time.Minute)).ToNot(HaveOccurred())

			key := types.NamespacedName{
				Namespace: namespaces.TestingNamespace.Name,
				Name:      sharedPodName,
			}
			sharedPod := &corev1.Pod{}
			err = testclient.Client.Get(context.TODO(), key, sharedPod)
			Expect(err).ToNot(HaveOccurred())

			By("inspecting the pod's accessible cpus")
			cpus, err := getCpusFrom(testclient.K8sClient, pod)
			Expect(err).ToNot(HaveOccurred())

			sharedCpus, err := getCpusFrom(testclient.K8sClient, sharedPod)
			Expect(err).ToNot(HaveOccurred())
			intersect := cpus.Intersection(*sharedCpus)
			Expect(intersect.Equals(*sharedCpus)).To(BeTrue(), "shared cpus: %s, are not present under pod: %v", sharedCpus.String(), fmt.Sprintf("%s/%s", pod.Namespace, pod.Name))

			By("inspecting the second pod's accessible cpus")
			cpus, err = getCpusFrom(testclient.K8sClient, pod2)
			Expect(err).ToNot(HaveOccurred())

			intersect = cpus.Intersection(*sharedCpus)
			Expect(intersect.Equals(*sharedCpus)).To(BeTrue(), "shared cpus: %s, are not present under pod: %v", sharedCpus.String(), fmt.Sprintf("%s/%s", pod2.Namespace, pod2.Name))

			By(fmt.Sprintf("deleting the test pod: %s", pod.Name))
			err = testclient.Client.Delete(context.TODO(), pod)
			Expect(err).ToNot(HaveOccurred())
			Expect(pods.WaitForDeletion(pod, time.Minute)).ToNot(HaveOccurred())

			By(fmt.Sprintf("deleting the second test pod: %s", pod2.Name))
			err = testclient.Client.Delete(context.TODO(), pod2)
			Expect(err).ToNot(HaveOccurred())
			Expect(pods.WaitForDeletion(pod2, time.Minute)).ToNot(HaveOccurred())
		})

		When("a pods gets created without the shared annotations", func() {
			It("should not contains the shared cpus", func() {
				pod := pods.GetTestPod(
					pods.WithResources(corev1.ResourceRequirements{
						Limits: corev1.ResourceList{
							corev1.ResourceCPU:    resource.MustParse("2"),
							corev1.ResourceMemory: resource.MustParse("100Mi"),
						},
					}))
				err := testclient.Client.Create(context.TODO(), pod)
				Expect(err).ToNot(HaveOccurred())
				Expect(pods.WaitForPhase(pod, corev1.PodRunning, time.Minute)).ToNot(HaveOccurred())

				key := types.NamespacedName{
					Namespace: namespaces.TestingNamespace.Name,
					Name:      sharedPodName,
				}
				sharedPod := &corev1.Pod{}
				err = testclient.Client.Get(context.TODO(), key, sharedPod)
				Expect(err).ToNot(HaveOccurred())

				By("inspecting the pod's accessible cpus")
				cpus, err := getCpusFrom(testclient.K8sClient, pod)
				Expect(err).ToNot(HaveOccurred())

				sharedCpus, err := getCpusFrom(testclient.K8sClient, sharedPod)
				Expect(err).ToNot(HaveOccurred())
				intersect := cpus.Intersection(*sharedCpus)
				Expect(intersect.IsEmpty()).To(BeTrue(), "shared cpus: %s, are present under pod: %v", sharedCpus.String(), fmt.Sprintf("%s/%s", pod.Namespace, pod.Name))

				By(fmt.Sprintf("deleting the test pod: %s", pod.Name))
				err = testclient.Client.Delete(context.TODO(), pod)
				Expect(err).ToNot(HaveOccurred())
				Expect(pods.WaitForDeletion(pod, time.Minute)).ToNot(HaveOccurred())
			})
		})
	})
})

func getCpusFrom(c *kubernetes.Clientset, pod *corev1.Pod) (*cpuset.CPUSet, error) {
	cmd := []string{"/bin/sh", "-c", "grep Cpus_allowed_list /proc/self/status | cut -f2"}
	out, err := pods.ExecCommandOnPod(c, pod, cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to run command: %s out: %s; %w", cmd, out, err)
	}

	cpus, err := cpuset.Parse(strings.Trim(string(out), "\r\n"))
	if err != nil {
		return nil, fmt.Errorf("failed to parse cpuset when input is: %s; %w", out, err)
	}
	return &cpus, nil
}
