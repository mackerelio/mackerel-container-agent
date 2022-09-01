package config

import (
	"context"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/mackerelio/mackerel-container-agent/cmdutil"
)

func TestLoaderLoad(t *testing.T) {
	file := newConfigFile(t, `
apikey: 'DUMMY APIKEY'
root: '/tmp/mackerel-container-agent'
`)

	expect := &Config{
		Apibase: "",
		Apikey:  "DUMMY APIKEY",
		Root:    "/tmp/mackerel-container-agent",
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	confLoader := NewLoader(file, 0)
	conf, err := confLoader.Load(ctx)
	if err != nil {
		t.Fatalf("should not raise error: %v", err)
	}
	if !reflect.DeepEqual(conf, expect) {
		t.Errorf("expect %#v, got %#v", expect, conf)
	}

	confCh := confLoader.Start(ctx)
	go cancel()
	<-confCh
}

func TestLoaderStart(t *testing.T) {
	file := newConfigFile(t, `
apikey: 'DUMMY APIKEY'
root: '/tmp/mackerel-container-agent'
`)

	expect := &Config{
		Apibase: "",
		Apikey:  "DUMMY APIKEY",
		Root:    "/tmp/mackerel-container-agent",
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	confLoader := NewLoader(file, 300*time.Millisecond)
	conf, err := confLoader.Load(ctx)
	if err != nil {
		t.Fatalf("should not raise error: %v", err)
	}
	if !reflect.DeepEqual(conf, expect) {
		t.Errorf("expect %#v, got %#v", expect, conf)
	}

	confCh := confLoader.Start(ctx)
	go func() {
		for {
			select {
			case <-confCh:
				cancel()
				return
			case <-ctx.Done():
				return
			}
		}
	}()

	errCh := make(chan error)
	go func() {
		time.Sleep(800 * time.Millisecond)
		errCh <- os.WriteFile(file, []byte(`
apikey: 'DUMMY APIKEY 2'
root: '/tmp/mackerel-container-agent'
plugin:
  metrics:
    mysql:
      command: mackerel-plugin-mysql
`), 0600)
	}()

	expect2 := &Config{
		Apibase: "",
		Apikey:  "DUMMY APIKEY 2",
		Root:    "/tmp/mackerel-container-agent",
		MetricPlugins: []*MetricPlugin{
			&MetricPlugin{
				Name:    "mysql",
				Command: cmdutil.CommandString("mackerel-plugin-mysql"),
			},
		},
	}

	<-ctx.Done()
	if err := <-errCh; err != nil {
		t.Fatalf("should not raise error (failed to write new config file): %v", err)
	}

	conf, err = confLoader.Load(ctx)
	if err != nil {
		t.Fatalf("should not raise error: %v", err)
	}
	if !reflect.DeepEqual(conf, expect2) {
		t.Errorf("expect %#v, got %#v", expect2, conf)
	}
}

func TestLoaderStartCancel(t *testing.T) {
	file := newConfigFile(t, `
apikey: 'DUMMY APIKEY'
root: '/tmp/mackerel-container-agent'
`)

	expect := &Config{
		Apibase: "",
		Apikey:  "DUMMY APIKEY",
		Root:    "/tmp/mackerel-container-agent",
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	confLoader := NewLoader(file, 300*time.Millisecond)
	conf, err := confLoader.Load(ctx)
	if err != nil {
		t.Fatalf("should not raise error: %v", err)
	}
	if !reflect.DeepEqual(conf, expect) {
		t.Errorf("expect %#v, got %#v", expect, conf)
	}

	confCh := confLoader.Start(ctx)

	go func() {
		time.Sleep(500 * time.Millisecond)
		cancel()
	}()

	<-confCh // when the context is done, loader should stop the polling loop
	<-ctx.Done()
}
