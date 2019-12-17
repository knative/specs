/*
Copyright 2019 The Knative Authors
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

// Package adapter implements a sample receive adapter that generates events
// at a regular interval.
package adapter

import (
	"context"
	cloudevents "github.com/cloudevents/sdk-go"
	"github.com/cloudevents/sdk-go/pkg/cloudevents/types"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"knative.dev/eventing/pkg/adapter"
	"knative.dev/pkg/logging"
	"knative.dev/pkg/source"
	"net/url"
	"time"
)

type envConfig struct {
	// Include the standard adapter.EnvConfig used by all adapters.
	adapter.EnvConfig

	// Interval between events, for example "5s", "100ms"
	Interval time.Duration `envconfig:"INTERVAL" required:"true"`
}

func NewEnv() adapter.EnvConfigAccessor { return &envConfig{} }

// Adapter generates events at a regular interval.
type Adapter struct {
	logger   *zap.Logger
	interval time.Duration
	nextID   int
	client   cloudevents.Client
}

type dataExample struct {
	Sequence  int
	Heartbeat string
}

var sourceURI = types.URIRef{URL: url.URL{Scheme: "http", Host: "sample.knative.dev", Path: "/heartbeat-source"}}

func strptr(s string) *string { return &s }

func (a *Adapter) newEvent() cloudevents.Event {

	e := cloudevents.Event{
		Context: cloudevents.EventContextV1{
			ID:              uuid.New().String(),
			Type:            "dev.knative.sample",
			Source:          sourceURI,
			Time:            &types.Timestamp{Time: time.Now()},
			DataContentType: strptr("application/json"),
		}.AsV1(),
		Data: &dataExample{
			Sequence:  a.nextID,
			Heartbeat: a.interval.String(),
		},
	}
	a.nextID++
	return e
}

// Start runs the adapter.
// Returns if stopCh is closed or Send() returns an error.
func (a *Adapter) Start(stopCh <-chan struct{}) error {
	a.logger.Info("Starting with: ",
		zap.String("Interval: ", a.interval.String()))
	for {
		select {
		case <-time.After(a.interval):
			event := a.newEvent()
			a.logger.Info("Sending new event: ", zap.String("event", event.String()))
			_, _, err := a.client.Send(context.Background(), event)
			if err != nil {
				return err
			}
		case <-stopCh:
			a.logger.Info("Shutting down...")
			return nil
		}
	}
}

func NewAdapter(ctx context.Context, aEnv adapter.EnvConfigAccessor, ceClient cloudevents.Client, reporter source.StatsReporter) adapter.Adapter {
	env := aEnv.(*envConfig) // Will always be our own envConfig type
	logger := logging.FromContext(ctx).Desugar()
	logger.Info("Heartbeat example", zap.Duration("interval", env.Interval))
	return &Adapter{
		interval: env.Interval,
		client:   ceClient,
		logger:   logger,
	}
}
