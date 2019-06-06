package csrx

import (
	"context"
	"reflect"

	commonv1alpha1 "github.com/michaelhenkel/csrx-operator/pkg/apis/common/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var log = logf.Log.WithName("controller_csrx")

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new Csrx Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileCsrx{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("csrx-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource Csrx
	err = c.Watch(&source.Kind{Type: &commonv1alpha1.Csrx{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// TODO(user): Modify this to be the types you create that are owned by the primary resource
	// Watch for changes to secondary resource Pods and requeue the owner Csrx
	err = c.Watch(&source.Kind{Type: &corev1.Pod{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &commonv1alpha1.Csrx{},
	})
	if err != nil {
		return err
	}

	return nil
}

// blank assignment to verify that ReconcileCsrx implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileCsrx{}

// ReconcileCsrx reconciles a Csrx object
type ReconcileCsrx struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a Csrx object and makes changes based on the state read
// and what is in the Csrx.Spec
// TODO(user): Modify this Reconcile function to implement your Controller logic.  This example creates
// a Pod as an example
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileCsrx) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling Csrx")

	// Fetch the Csrx instance
	instance := &commonv1alpha1.Csrx{}
	err := r.client.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	// Define a new Pod object
	pod := r.newPodForCR(instance)

	// Set Csrx instance as the owner and controller
	if err := controllerutil.SetControllerReference(instance, pod, r.scheme); err != nil {
		return reconcile.Result{}, err
	}


	err = r.CreateEmptyConfigMap(instance)
	if err != nil {
		return reconcile.Result{}, err
	}


	// Check if this Pod already exists
	found := &corev1.Pod{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: pod.Name, Namespace: pod.Namespace}, found)
	if err != nil && errors.IsNotFound(err) {
		reqLogger.Info("Creating a new Pod", "Pod.Namespace", pod.Namespace, "Pod.Name", pod.Name)
		err = r.client.Create(context.TODO(), pod)
		if err != nil {
			return reconcile.Result{}, err
		}

		// Pod created successfully - don't requeue
		return reconcile.Result{}, nil
	} else if err != nil {
		return reconcile.Result{}, err
	}

	var podNames []string
	podNames, err = r.GetPodNames(instance)
	if err != nil {
		reqLogger.Error(err, "Failed to get PodNames")
		return reconcile.Result{}, err
	} else {
		reqLogger.Info("Got PodNames")
	}

	if !reflect.DeepEqual(podNames, instance.Status.Nodes) {
		instance.Status.Nodes = podNames
		err = r.client.Status().Update(context.TODO(), instance)
		if err != nil {
			reqLogger.Error(err, "Failed to update Pod status.")
			return reconcile.Result{}, err
		}
	}
	reqLogger.Info("Updated Node status with PodNames")

	// Update ConfigMap
	err = r.UpdateConfigMap(instance)
	if err != nil {
		return reconcile.Result{}, err
	}

	reqLogger.Info("Config configmap updated")

	reqLogger.Info("Skip reconcile: Pod already exists", "Pod.Namespace", found.Namespace, "Pod.Name", found.Name)
	return reconcile.Result{}, nil
}

func (c *ReconcileCsrx) DeleteInitConfigMap(cr *commonv1alpha1.Csrx) (error) {
	initConfigMap := &corev1.ConfigMap{}
	err := c.client.Get(context.TODO(), types.NamespacedName{Name: cr.Name + "-interfaces", Namespace: cr.Namespace}, initConfigMap)
	if err != nil && errors.IsNotFound(err) {
		return nil
	}
	err = c.client.Delete(context.TODO(), initConfigMap)
	if err != nil {
		return err
	}
	return nil
}

func (c *ReconcileCsrx) SetPrefixFromConfigMap(cr *commonv1alpha1.Csrx) (error) {
	initConfigMap := &corev1.ConfigMap{}
	err := c.client.Get(context.TODO(), types.NamespacedName{Name: cr.Name + "-init-cm", Namespace: cr.Namespace}, initConfigMap)
	if err != nil && errors.IsNotFound(err) {
		return err
	}
	if prefix, ok := initConfigMap.Data["prefix"]; ok {
		cr.Status.Prefix = prefix
		err = c.client.Status().Update(context.TODO(), cr)
		if err != nil {
			return err
		}
	} else {
		return errors.NewBadRequest("Failed to get prefix from config map")
	}
	return nil

}

func (c *ReconcileCsrx) CreateEmptyConfigMap(cr *commonv1alpha1.Csrx) (error) {
	
	configMap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name: cr.Name + "-cm",
			Namespace: cr.Namespace,
		},
		Data: map[string]string{},
	}

	err := c.client.Get(context.TODO(), types.NamespacedName{Name: cr.Name + "-cm", Namespace: cr.Namespace}, configMap)
	if err != nil && errors.IsNotFound(err) {
		controllerutil.SetControllerReference(cr, configMap, c.scheme)
		err = c.client.Create(context.TODO(), configMap)
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *ReconcileCsrx) UpdateConfigMap(cr *commonv1alpha1.Csrx) (error) {
	

	interfaceConfigMap := &corev1.ConfigMap{}
	err := c.client.Get(context.TODO(), types.NamespacedName{Name: cr.Name + "-interfaces", Namespace: cr.Namespace}, interfaceConfigMap)
	if err != nil && errors.IsNotFound(err) {
		return err
	}
	controllerutil.SetControllerReference(cr, interfaceConfigMap, c.scheme)
	err = c.client.Update(context.TODO(), interfaceConfigMap)
	if err != nil {
		return err
	}

	interfaceMap := interfaceConfigMap.Data

	var junosConfigString string
	if len(interfaceMap) > 1 {
		if address, ok := interfaceMap["eth1"]; ok{
			junosConfigString = junosConfigString +`
set interfaces ge-0/0/1 unit 0 family inet address `+ address + `
`	
		}

		if address, ok := interfaceMap["eth2"]; ok{
			junosConfigString = junosConfigString +`
set interfaces ge-0/0/0 unit 0 family inet address `+ address + `
`	
		}

	}
/*
	junosConfigString = junosConfigString +`
set interfaces ge-0/0/1 unit 0 family inet address 1.1.1.1/24
`

	junosConfigString = junosConfigString +`
set interfaces ge-0/0/0 unit 0 family inet address 1.1.0.1/24
`
*/
	var resourceConfig = make(map[string]string)
	resourceConfig["junosconfig"] = junosConfigString

	initConfigMap := &corev1.ConfigMap{}
	err = c.client.Get(context.TODO(), types.NamespacedName{Name: cr.Name + "-cm", Namespace: cr.Namespace}, initConfigMap)
	if err != nil && errors.IsNotFound(err) {
		return err
	} else {
		initConfigMap.Data = resourceConfig
		err = c.client.Update(context.TODO(), initConfigMap)
		if err != nil {
			return err
		}
	}

	return nil
}

// newPodForCR returns a busybox pod with the same name/namespace as the cr
func (c *ReconcileCsrx)newPodForCR(cr *commonv1alpha1.Csrx) *corev1.Pod {
	var imagePullSecretsList []corev1.LocalObjectReference
	for _, imagePullSecretName := range(cr.Spec.ImagePullSecrets){
		imagePullSecret := corev1.LocalObjectReference{
			Name: imagePullSecretName,
		}
		imagePullSecretsList = append(imagePullSecretsList, imagePullSecret)
	}
	var volumeList []corev1.Volume
	statusVolume := corev1.Volume{
		Name: "status",
		VolumeSource: corev1.VolumeSource{
			DownwardAPI: &corev1.DownwardAPIVolumeSource{
				Items: []corev1.DownwardAPIVolumeFile{
					corev1.DownwardAPIVolumeFile{
						Path: "pod_labels",
						FieldRef: &corev1.ObjectFieldSelector{
							FieldPath: "metadata.labels",
						},
					},
				},
			},
		},
	}
	junosConfigVolume := corev1.Volume{
		Name: "junosconfig",
		VolumeSource: corev1.VolumeSource{
			ConfigMap: &corev1.ConfigMapVolumeSource{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: cr.Name + "-cm",
				},
			},
		},
	}

	configVolume := corev1.Volume{
		Name: "config",
		VolumeSource: corev1.VolumeSource{
			EmptyDir: &corev1.EmptyDirVolumeSource{},
		},
	}

	volumeList = append(volumeList, statusVolume)
	volumeList = append(volumeList, junosConfigVolume)
	volumeList = append(volumeList, configVolume)

	var annotationsMap = make(map[string]string) 
	if len(cr.Spec.Networks) > 0 {
		nwString := "["
		for idx, nw := range(cr.Spec.Networks){
			nwString = nwString +"{"
			if nw.Name != "" {
				nwString = nwString + "\"name\": \"" + nw.Name + "\""
			}
			if nw.Interface != "" {
				nwString = nwString + ", \"interface\": \"" + nw.Interface + "\""
			}
			if nw.Namespace != "" {
				nwString = nwString + ", \"namespace\": \"" + nw.Namespace + "\""
			}
			nwString = nwString +"}"
			if idx+1 != len(cr.Spec.Networks){
				nwString = nwString +","
			}
		}
		nwString = nwString + "]"
		annotationsMap["k8s.v1.cni.cncf.io/networks"] = nwString
	}


	privileged := true
	labels := map[string]string{
		"app": cr.Name,
	}
	csrxInitCommand := []string{
		"/csrx-init",
		cr.Name + "-interfaces",
	}

	csrxCommand := []string{
		"bash",
		"-c",
		"while IFS= read -r line; do sed -i \"/set security zones security-zone untrust/a $line\" /etc/rc.local; done < <(cat /etc/csrxconfig/junosconfig); /etc/rc.local init",
	}

	var initImagePullPolicy string
	if cr.Spec.InitImagePullPolicy != "" {
		initImagePullPolicy = "Always" 
	} else {
		initImagePullPolicy = cr.Spec.InitImagePullPolicy
	}

	var csrxImagePullPolicy string
	if cr.Spec.CsrxImagePullPolicy != "" {
		csrxImagePullPolicy = "Always" 
	} else {
		csrxImagePullPolicy = cr.Spec.CsrxImagePullPolicy
	}

	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Name + "-pod",
			Namespace: cr.Namespace,
			Labels:    labels,
			Annotations: annotationsMap,
		},
		Spec: corev1.PodSpec{
			ServiceAccountName: "csrx-operator",
			InitContainers: []corev1.Container{
				{
					Name:    "csrx-init",
					ImagePullPolicy: corev1.PullPolicy(initImagePullPolicy),
					Image:   cr.Spec.InitImage,
					Command: csrxInitCommand,
				},
			},
			Containers: []corev1.Container{
				{
					Name:    "csrx",
					ImagePullPolicy: corev1.PullPolicy(csrxImagePullPolicy),
					Image:   cr.Spec.CsrxImage,
					Command: csrxCommand,
					SecurityContext: &corev1.SecurityContext{
						Privileged: &privileged,
					},
					VolumeMounts: []corev1.VolumeMount{{
						Name: "junosconfig",
						ReadOnly: false,
						MountPath: "/etc/csrxconfig",
					}},
				},
			},
			Volumes: volumeList,
			ImagePullSecrets: imagePullSecretsList,
		},
	}
}

func (c *ReconcileCsrx) IsInitContainerReady(cr *commonv1alpha1.Csrx, podName string) (bool) {
	foundPod := &corev1.Pod{}
	err := c.client.Get(context.TODO(), types.NamespacedName{Name: podName, Namespace: cr.Namespace}, foundPod)
	if err != nil {
		return false
	}
	podMetaData := foundPod.ObjectMeta
	podLabels := podMetaData.Labels
	if podLabels["status"] == "ready" {
		return true
	}
	return false
}

func (c *ReconcileCsrx) GetPodNames(cr *commonv1alpha1.Csrx) ([]string, error) {
	podList := &corev1.PodList{}
	labelSelector := labels.SelectorFromSet(map[string]string{"app": cr.Name})
	var podNames []string

	listOps := &client.ListOptions{
		Namespace:     cr.Namespace,
		LabelSelector: labelSelector,
	}
	err := c.client.List(context.TODO(), listOps, podList)
	if err != nil {
		return podNames, err
	}
	for _, pod := range podList.Items {
		podNames = append(podNames, pod.Name)
	}
	return podNames, nil
}

func (c *ReconcileCsrx) LabelPod(cr *commonv1alpha1.Csrx, podName string) (*corev1.Pod, error) {
	foundPod := &corev1.Pod{}
	err := c.client.Get(context.TODO(), types.NamespacedName{Name: podName, Namespace: cr.Namespace}, foundPod)
	if err != nil {
		return foundPod, err
	}
	podMetaData := foundPod.ObjectMeta
	podLabels := podMetaData.Labels
	podLabels["status"] = "ready"
	foundPod.ObjectMeta.Labels = podLabels
	return foundPod, nil
}
