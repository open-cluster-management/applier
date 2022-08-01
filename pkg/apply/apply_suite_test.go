// Copyright Contributors to the Open Cluster Management project

package apply

import (
	"context"
	"path/filepath"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apiextensionsclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/kubectl/pkg/scheme"
	"sigs.k8s.io/controller-runtime/pkg/envtest"

	"github.com/stolostron/applier/test/unit/resources/scenario"
)

var testEnv *envtest.Environment
var restConfig *rest.Config
var kubeClient kubernetes.Interface
var apiExtensionsClient apiextensionsclient.Interface
var dynamicClient dynamic.Interface
var GvrSCR schema.GroupVersionResource = schema.GroupVersionResource{Group: "example.com", Version: "v1", Resource: "samplecustomresources"}

func TestApply(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "TemplateFunction Suite")
}

var _ = BeforeSuite(func() {
	By("bootstrapping test environment")

	// start a kube-apiserver
	testEnv = &envtest.Environment{
		ErrorIfCRDPathMissing: true,
		CRDDirectoryPaths: []string{
			filepath.Join("..", "..", "test", "unit", "resources", "scenario", "config", "crd", "crd.yaml"),
		},
	}

	cfg, err := testEnv.Start()
	Expect(err).ToNot(HaveOccurred())
	Expect(cfg).ToNot(BeNil())

	kubeClient, err = kubernetes.NewForConfig(cfg)
	Expect(err).NotTo(HaveOccurred())
	apiExtensionsClient, err = apiextensionsclient.NewForConfig(cfg)
	Expect(err).NotTo(HaveOccurred())
	dynamicClient, err = dynamic.NewForConfig(cfg)
	Expect(err).NotTo(HaveOccurred())

	restConfig = cfg
})

var _ = AfterSuite(func() {
	By("tearing down the test environment")

	err := testEnv.Stop()
	Expect(err).ToNot(HaveOccurred())
})

var _ = Describe("setOwnerRef", func() {
	It("Add OwnerRef to core item", func() {
		var nsOwner *corev1.Namespace
		By("Creating ns owner", func() {
			nsOwner = &corev1.Namespace{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "v1",
					Kind:       "Namespace",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: "my-ns-owner-1",
				},
			}
			var err error
			nsOwner, err = kubeClient.CoreV1().Namespaces().Create(context.TODO(), nsOwner, metav1.CreateOptions{})
			Expect(err).To(BeNil())
		})
		By("setReferenceOwner", func() {
			reader := scenario.GetScenarioResourcesReader()
			applierBuilder := NewApplierBuilder()
			applier := applierBuilder.
				WithClient(kubeClient, apiExtensionsClient, dynamicClient).
				WithTemplateFuncMap(FuncMap()).WithOwner(nsOwner, false, false, scheme.Scheme).
				Build()

			_, err := applier.ApplyDirectly(reader, nil, false, "", "ownerref/ns.yaml")
			Expect(err).To(BeNil())
		})
		By("Checking Ownerref", func() {
			ns, err := kubeClient.CoreV1().Namespaces().Get(context.TODO(), "my-ns", metav1.GetOptions{})
			Expect(err).To(BeNil())
			Expect(len(ns.GetOwnerReferences())).To(Equal(1))
			Expect(ns.OwnerReferences[0].APIVersion).To(Equal(nsOwner.APIVersion))
			Expect(ns.OwnerReferences[0].Kind).To(Equal(nsOwner.Kind))
			Expect(ns.OwnerReferences[0].Name).To(Equal(nsOwner.Name))
			Expect(ns.OwnerReferences[0].UID).To(Equal(nsOwner.UID))
			Expect(ns.OwnerReferences[0].Controller).To(BeNil())
			Expect(ns.OwnerReferences[0].BlockOwnerDeletion).To(BeNil())
		})
	})
	It("Add OwnerRef to CRD item", func() {
		var sampleOwner *unstructured.Unstructured
		By("Creating sample owner", func() {
			object := make(map[string]interface{})
			object["metadata"] = metav1.ObjectMeta{
				Name: "my-sampleowner-owner-1",
			}
			sampleOwner = &unstructured.Unstructured{
				Object: object,
			}
			sampleOwner.SetAPIVersion("example.com/v1")
			sampleOwner.SetKind("SampleCustomResource")
			var err error
			sampleOwner, err = dynamicClient.Resource(GvrSCR).Create(context.TODO(), sampleOwner, metav1.CreateOptions{})
			Expect(err).To(BeNil())
		})
		By("setReferenceOwner", func() {
			reader := scenario.GetScenarioResourcesReader()
			applierBuilder := NewApplierBuilder()
			applier := applierBuilder.
				WithClient(kubeClient, apiExtensionsClient, dynamicClient).
				WithTemplateFuncMap(FuncMap()).WithOwner(sampleOwner, false, false, scheme.Scheme).
				Build()

			_, err := applier.ApplyCustomResources(reader, nil, false, "", "ownerref/sampleowner.yaml")
			Expect(err).To(BeNil())
		})
		By("Checking Ownerref", func() {
			sample, err := dynamicClient.Resource(GvrSCR).Get(context.TODO(), "my-sampleowner", metav1.GetOptions{})
			Expect(err).To(BeNil())
			Expect(len(sample.GetOwnerReferences())).To(Equal(1))
			Expect(sample.GetOwnerReferences()[0].APIVersion).To(Equal(sampleOwner.GetAPIVersion()))
			Expect(sample.GetOwnerReferences()[0].Kind).To(Equal(sampleOwner.GetKind()))
			Expect(sample.GetOwnerReferences()[0].Name).To(Equal(sampleOwner.GetName()))
			Expect(sample.GetOwnerReferences()[0].UID).To(Equal(sampleOwner.GetUID()))
			Expect(sample.GetOwnerReferences()[0].Controller).To(BeNil())
			Expect(sample.GetOwnerReferences()[0].BlockOwnerDeletion).To(BeNil())
		})
	})
	It("Add OwnerRef to Deployment item", func() {
		var deployment *appsv1.Deployment
		By("Creating cluster owner", func() {
			reader := scenario.GetScenarioResourcesReader()
			applierBuilder := NewApplierBuilder()
			applier := applierBuilder.
				WithClient(kubeClient, apiExtensionsClient, dynamicClient).
				WithTemplateFuncMap(FuncMap()).
				Build()

			values := struct {
				Name string
			}{
				Name: "my-deployment-owner",
			}
			_, err := applier.ApplyCustomResources(reader, values, false, "", "ownerref/deployment.yaml")
			Expect(err).To(BeNil())
			deployment, err = kubeClient.AppsV1().Deployments("my-ns").Get(context.TODO(), "my-deployment-owner", metav1.GetOptions{})
			Expect(err).To(BeNil())
		})
		By("setReferenceOwner", func() {
			reader := scenario.GetScenarioResourcesReader()
			applierBuilder := NewApplierBuilder()
			applier := applierBuilder.
				WithClient(kubeClient, apiExtensionsClient, dynamicClient).
				WithTemplateFuncMap(FuncMap()).WithOwner(deployment, false, false, scheme.Scheme).
				Build()

			values := struct {
				Name string
			}{
				Name: "my-deployment",
			}
			_, err := applier.ApplyCustomResources(reader, values, false, "", "ownerref/deployment.yaml")
			Expect(err).To(BeNil())
		})
		By("Checking Ownerref", func() {
			dep, err := kubeClient.AppsV1().Deployments("my-ns").Get(context.TODO(), "my-deployment", metav1.GetOptions{})
			Expect(err).To(BeNil())
			Expect(len(dep.GetOwnerReferences())).To(Equal(1))
			Expect(dep.OwnerReferences[0].APIVersion).To(Equal(deployment.APIVersion))
			Expect(dep.OwnerReferences[0].Kind).To(Equal(deployment.Kind))
			Expect(dep.OwnerReferences[0].Name).To(Equal(deployment.Name))
			Expect(dep.OwnerReferences[0].UID).To(Equal(deployment.UID))
			Expect(dep.OwnerReferences[0].Controller).To(BeNil())
			Expect(dep.OwnerReferences[0].BlockOwnerDeletion).To(BeNil())
		})
	})
	It("Add OwnerRef to Deployment item with controller and blockDeletion", func() {
		var deployment *appsv1.Deployment
		By("Creating cluster owner", func() {
			reader := scenario.GetScenarioResourcesReader()
			applierBuilder := NewApplierBuilder()
			applier := applierBuilder.
				WithClient(kubeClient, apiExtensionsClient, dynamicClient).
				WithTemplateFuncMap(FuncMap()).
				Build()

			values := struct {
				Name string
			}{
				Name: "my-deployment-owner-controller",
			}
			_, err := applier.ApplyCustomResources(reader, values, false, "", "ownerref/deployment.yaml")
			Expect(err).To(BeNil())
			deployment, err = kubeClient.AppsV1().Deployments("my-ns").Get(context.TODO(), "my-deployment-owner-controller", metav1.GetOptions{})
			Expect(err).To(BeNil())
		})
		By("setReferenceOwner", func() {
			reader := scenario.GetScenarioResourcesReader()
			applierBuilder := NewApplierBuilder()
			applier := applierBuilder.
				WithClient(kubeClient, apiExtensionsClient, dynamicClient).
				WithTemplateFuncMap(FuncMap()).WithOwner(deployment, true, true, scheme.Scheme).
				Build()

			values := struct {
				Name string
			}{
				Name: "my-deployment-controller",
			}
			_, err := applier.ApplyCustomResources(reader, values, false, "", "ownerref/deployment.yaml")
			Expect(err).To(BeNil())
		})
		By("Checking Ownerref", func() {
			dep, err := kubeClient.AppsV1().Deployments("my-ns").Get(context.TODO(), "my-deployment-controller", metav1.GetOptions{})
			Expect(err).To(BeNil())
			Expect(len(dep.GetOwnerReferences())).To(Equal(1))
			Expect(dep.OwnerReferences[0].APIVersion).To(Equal(deployment.APIVersion))
			Expect(dep.OwnerReferences[0].Kind).To(Equal(deployment.Kind))
			Expect(dep.OwnerReferences[0].Name).To(Equal(deployment.Name))
			Expect(dep.OwnerReferences[0].UID).To(Equal(deployment.UID))
			Expect(*dep.OwnerReferences[0].Controller).To(BeTrue())
			Expect(*dep.OwnerReferences[0].BlockOwnerDeletion).To(BeTrue())
		})
	})
})
