package resources

import (
	"context"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	sourcesv1 "knative.dev/eventing/pkg/apis/sources/v1"
	duckv1 "knative.dev/pkg/apis/duck/v1"
	"knative.dev/pkg/kmeta"
	"knative.dev/pkg/tracker"
)

func SinkBindingName(source, subject string) string {
	return kmeta.ChildName(fmt.Sprintf("%s-%s", source, subject), "-sinkbinding")
}

func MakeSinkBinding(owner kmeta.OwnerRefable, source duckv1.SourceSpec, subject tracker.Reference) *sourcesv1.SinkBinding {
	sb := &sourcesv1.SinkBinding{
		ObjectMeta: metav1.ObjectMeta{
			OwnerReferences: []metav1.OwnerReference{
				*kmeta.NewControllerRef(owner),
			},
			Name:      SinkBindingName(owner.GetObjectMeta().GetName(), subject.Name),
			Namespace: owner.GetObjectMeta().GetNamespace(),
		},
		Spec: sourcesv1.SinkBindingSpec{
			SourceSpec: source,
			BindingSpec: duckv1.BindingSpec{
				Subject: subject,
			},
		},
	}

	sb.SetDefaults(context.Background())
	return sb
}
