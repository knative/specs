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

	ce "github.com/cloudevents/sdk-go/v1"
	cloudevents "github.com/cloudevents/sdk-go/v1"
	"github.com/cloudevents/sdk-go/v1/cloudevents/transport"
	cehttp "github.com/cloudevents/sdk-go/v1/cloudevents/transport/http"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"knative.dev/eventing/pkg/adapter"
	"knative.dev/eventing/pkg/kncloudevents"
	"knative.dev/pkg/logging"
)

func TestAdapter(t *testing.T) {
	// Test sink to receive events.
	sink := newSink(t)
	defer sink.close()
	c, err := kncloudevents.NewDefaultClient(sink.URL())
	require.NoError(t, err)

	// Keep the adapter logging quiet for tests.
	ctx := logging.WithLogger(context.Background(), zap.NewNop().Sugar())
	a := NewAdapter(ctx, &envConfig{Interval: time.Duration(time.Millisecond)}, c, nil)
	stop := make(chan struct{})
	go a.Start(stop)
	defer func() { close(stop) }()
	verify(t, sink.received)
}

func verify(t *testing.T, received chan ce.Event) {
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
		"SINK_URI="+sink.URL(),
		"INTERVAL="+"1ms",
		"NAMESPACE=namespace",
		`K_METRICS_CONFIG={"domain":"x", "component":"x", "prometheusport":0, "configmap":{}}`,
		`K_LOGGING_CONFIG={}`,
	)
	cmd.Start()
	defer func() { cmd.Process.Kill(); cmd.Wait() }()
	verify(t, sink.received)
}

type sink struct {
	listener  net.Listener
	transport transport.Transport
	ctx       context.Context
	close     func()
	received  chan ce.Event
}

func newSink(t *testing.T) *sink {
	s := &sink{received: make(chan ce.Event)}
	s.ctx, s.close = context.WithCancel(context.Background())
	var err error
	s.listener, err = net.Listen("tcp", ":0")
	require.NoError(t, err)
	s.transport, err = cehttp.New(cehttp.WithListener(s.listener))
	s.transport.SetReceiver(transport.ReceiveFunc(
		func(ctx context.Context, e cloudevents.Event, _ *cloudevents.EventResponse) error {
			select {
			case s.received <- e:
				return nil
			case <-ctx.Done():
				return ctx.Err()
			}
		}))
	go func() {
		_ = s.transport.StartReceiver(s.ctx)
		close(s.received)
	}()
	return s
}

func (s *sink) URL() string { return "http://" + s.listener.Addr().String() }
