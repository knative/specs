package reconciler

import (
	"context"
	"fmt"

	// k8s.io imports
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes"

	// knative.dev imports
	duckv1 "knative.dev/pkg/apis/duck/v1"
	"knative.dev/pkg/kmeta"
	"knative.dev/pkg/logging"
	pkgreconciler "knative.dev/pkg/reconciler"

	sourcesv1 "knative.dev/eventing/pkg/apis/sources/v1"

	"go.uber.org/zap"
)

// newDeploymentCreated makes a new reconciler event with event type Normal, and
// reason DeploymentCreated.
func newDeploymentCreated(namespace, name string) pkgreconciler.Event {
	return pkgreconciler.NewEvent(corev1.EventTypeNormal, "DeploymentCreated", "created deployment: \"%s/%s\"", namespace, name)
}

// newDeploymentFailed makes a new reconciler event with event type Warning, and
// reason DeploymentFailed.
func newDeploymentFailed(namespace, name string, err error) pkgreconciler.Event {
	return pkgreconciler.NewEvent(corev1.EventTypeWarning, "DeploymentFailed", "failed to create deployment: \"%s/%s\", %w", namespace, name, err)
}

// newDeploymentUpdated makes a new reconciler event with event type Normal, and
// reason DeploymentUpdated.
func newDeploymentUpdated(namespace, name string) pkgreconciler.Event {
	return pkgreconciler.NewEvent(corev1.EventTypeNormal, "DeploymentUpdated", "updated deployment: \"%s/%s\"", namespace, name)
}

type DeploymentReconciler struct {
	KubeClientSet kubernetes.Interface
}

// ReconcileDeployment reconciles deployment resource for SampleSource
func (r *DeploymentReconciler) ReconcileDeployment(
	ctx context.Context,
	owner kmeta.OwnerRefable,
	binder *sourcesv1.SinkBinding,
	expected *appsv1.Deployment,
) (*appsv1.Deployment, *sourcesv1.SinkBinding, pkgreconciler.Event) {
	namespace := owner.GetObjectMeta().GetNamespace()
	ra, err := r.KubeClientSet.AppsV1().Deployments(namespace).Get(ctx, expected.Name, metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		syncSink(ctx, binder, expected.Spec.Template.Spec)
		ra, err = r.KubeClientSet.AppsV1().Deployments(namespace).Create(ctx, expected, metav1.CreateOptions{})
		if err != nil {
			return nil, binder, newDeploymentFailed(expected.Namespace, expected.Name, err)
		}
		return ra, binder, newDeploymentCreated(ra.Namespace, ra.Name)
	} else if err != nil {
		return nil, binder, fmt.Errorf("error getting receive adapter %q: %v", expected.Name, err)
	} else if !metav1.IsControlledBy(ra, owner.GetObjectMeta()) {
		return nil, binder, fmt.Errorf("deployment %q is not owned by %s %q",
			ra.Name, owner.GetGroupVersionKind().Kind, owner.GetObjectMeta().GetName())
	} else if podSpecSync(ctx, binder, expected.Spec.Template.Spec, ra.Spec.Template.Spec) {
		if ra, err = r.KubeClientSet.AppsV1().Deployments(namespace).Update(ctx, ra, metav1.UpdateOptions{}); err != nil {
			return ra, binder, err
		}
		return ra, binder, newDeploymentUpdated(ra.Namespace, ra.Name)
	} else {
		logging.FromContext(ctx).Debugw("Reusing existing receive adapter", zap.Any("receiveAdapter", ra))
	}
	return ra, binder, nil
}

func (r *DeploymentReconciler) FindOwned(ctx context.Context, owner kmeta.OwnerRefable, selector labels.Selector) (*appsv1.Deployment, error) {
	dl, err := r.KubeClientSet.AppsV1().Deployments(owner.GetObjectMeta().GetNamespace()).List(ctx, metav1.ListOptions{
		LabelSelector: selector.String(),
	})
	if err != nil {
		logging.FromContext(ctx).Error("Unable to list deployments: %v", zap.Error(err))
		return nil, err
	}
	for _, dep := range dl.Items {
		if metav1.IsControlledBy(&dep, owner.GetObjectMeta()) {
			return &dep, nil
		}
	}
	return nil, apierrors.NewNotFound(schema.GroupResource{}, "")
}

func getContainer(name string, spec corev1.PodSpec) (int, *corev1.Container) {
	for i, c := range spec.Containers {
		if c.Name == name {
			return i, &c
		}
	}
	return -1, nil
}

// Returns true if an update is needed.
func podSpecSync(ctx context.Context, binder *sourcesv1.SinkBinding, expected corev1.PodSpec, now corev1.PodSpec) bool {
	old := *now.DeepCopy()
	syncImage(expected, now)
	syncSink(ctx, binder, now)

	return !equality.Semantic.DeepEqual(old, now)
}

func syncSink(ctx context.Context, binder *sourcesv1.SinkBinding, now corev1.PodSpec) {
	// call Do() to project sink information.
	ps := &duckv1.WithPod{}
	ps.Spec.Template.Spec = now

	binder.Do(ctx, ps)
}

func syncImage(expected corev1.PodSpec, now corev1.PodSpec) {
	// got needs all of the containers that want as, but it is allowed to have more.
	for _, ec := range expected.Containers {
		n, nc := getContainer(ec.Name, now)
		if nc == nil {
			now.Containers = append(now.Containers, ec)
			continue
		}
		if nc.Image != ec.Image {
			now.Containers[n].Image = ec.Image
		}
	}
}
