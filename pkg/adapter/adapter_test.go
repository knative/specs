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

package adapter

import (
	"context"
	"net"
	"os"
	"os/exec"
	"testing"
	"time"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"knative.dev/eventing/pkg/adapter/v2"
	"knative.dev/pkg/logging"
)

func TestAdapter(t *testing.T) {
	// Test sink to receive events.
	sink := newSink(t)
	defer sink.close()

	tr, err := cloudevents.NewHTTP(cloudevents.WithTarget(sink.URL()))
	require.NoError(t, err)
	c, err := cloudevents.NewClient(tr, cloudevents.WithUUIDs())
	require.NoError(t, err)

	// Keep the adapter logging quiet for tests.
	ctx := logging.WithLogger(context.Background(), zap.NewNop().Sugar())
	a := NewAdapter(ctx, &envConfig{Interval: time.Millisecond}, c)
	ctx, cancel := context.WithCancel(ctx)
	go func() {
		if err := a.Start(ctx); err != nil {
			logging.FromContext(ctx).Errorw("failed to start adapter", zap.Error(err))
		}
	}()
	defer func() { cancel() }()
	verify(t, sink.received)
}

func verify(t *testing.T, received chan cloudevents.Event) {
	for _, id := range []int{0, 1, 2} {
		e := <-received
		assert.Equal(t, "dev.knative.sample", e.Type())
		//m := map[string]json.RawMessage{}
		m := &dataExample{}
		assert.NoError(t, e.DataAs(&m))
		n := &dataExample{Sequence: id, Heartbeat: "1ms"}
		assert.Equal(t, n, m)
	}
}

func TestAdapterMain(t *testing.T) {
	// Use the test executable to simulate the cmd/receive_adapter process if
	// environment var t.Name() is set to "main"
	// (see https://talks.golang.org/2014/testing.slide#23)
	if os.Getenv(t.Name()) == "main" {
		adapter.Main("sample-source", NewEnv, NewAdapter)
		return
	}

	// Set up a test sink to receive from the adapter.
	sink := newSink(t)
	defer sink.close()

	// Run a simulated receive_adapter main using the test executable.
	cmd := exec.Command(os.Args[0], "-test.run="+t.Name())
	cmd.Env = append(os.Environ(),
		t.Name()+"=main",
		"K_SINK="+sink.URL(),
		"INTERVAL="+"1ms",
		"NAMESPACE=namespace",
		"NAME=name",
		`K_METRICS_CONFIG={"domain":"x", "component":"x", "prometheusport":0, "configmap":{}}`,
		`K_LOGGING_CONFIG={}`,
	)
	err := cmd.Start()
	if err != nil {
		t.Error(err)
	}
	defer func() { cmd.Process.Kill(); cmd.Wait() }()
	verify(t, sink.received)
}

type sink struct {
	listener net.Listener
	client   cloudevents.Client
	proto    *cloudevents.HTTPProtocol
	ctx      context.Context
	close    func()
	received chan cloudevents.Event
}

func newSink(t *testing.T) *sink {
	s := &sink{received: make(chan cloudevents.Event)}
	//s.ctx, s.close = context.WithTimeout(context.Background(), 5*time.Second)
	s.ctx, s.close = context.WithTimeout(context.Background(), 1500*time.Millisecond)
	var err error
	s.listener, err = net.Listen("tcp", ":0")
	require.NoError(t, err)

	s.proto, err = cloudevents.NewHTTP(cloudevents.WithListener(s.listener))
	require.NoError(t, err)

	s.client, err = cloudevents.NewClient(s.proto)
	require.NoError(t, err)

	go func() {
		_ = s.client.StartReceiver(s.ctx, func(ctx context.Context, e cloudevents.Event) cloudevents.Result {
			select {
			case s.received <- e:
				return nil
			case <-ctx.Done():
				return ctx.Err()
			}
		})
		close(s.received)
	}()

	return s
}

func (s *sink) URL() string { return "http://" + s.listener.Addr().String() }
