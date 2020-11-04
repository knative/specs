/*
Copyright 2020 The Knative Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha1

import (
	"context"

	"knative.dev/pkg/apis"
)

// SetDefaults mutates SampleSource.
func (s *SampleSource) SetDefaults(ctx context.Context) {
	//Add code for Mutating admission webhook.

	//example: If ServiceAccountName is unspecified, default to the "default" service account.
	if s != nil && s.Spec.ServiceAccountName == "" {
		s.Spec.ServiceAccountName = "default"
	}

	//example: If Interval is unspecified, default to "10s".
	if s != nil && s.Spec.Interval == "" {
		s.Spec.Interval = "10s"
	}

	// call SetDefaults against duckv1.Destination with a context of ObjectMeta of SampleSource.
	withNS := apis.WithinParent(ctx, s.ObjectMeta)
	s.Spec.Sink.SetDefaults(withNS)
}
