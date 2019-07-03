package config

import (
	"context"
	"io/ioutil"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/mackerelio/mackerel-container-agent/cmdutil"
)

func TestLoaderLoad(t *testing.T) {
	file, err := newConfigFile(`
apikey: 'DUMMY APIKEY'
root: '/tmp/mackerel-container-agent'
`)
	if err != nil {
		t.Errorf("should not raise error: %v", err)
	}
	defer os.Remove(file.Name())

	expect := &Config{
		Apibase: "",
		Apikey:  "DUMMY APIKEY",
		Root:    "/tmp/mackerel-container-agent",
	}

	confLoader := NewLoader(file.Name(), 0)
	conf, err := confLoader.Load()
	if err != nil {
		t.Errorf("should not raise error: %v", err)
	}
	if !reflect.DeepEqual(conf, expect) {
		t.Errorf("expect %#v, got %#v", expect, conf)
	}

	confCh := confLoader.Start(context.Background())
	<-confCh // loader does not start the polling loop when duration is 0
}

func TestLoaderStart(t *testing.T) {
	file, err := newConfigFile(`
apikey: 'DUMMY APIKEY'
root: '/tmp/mackerel-container-agent'
`)
	if err != nil {
		t.Errorf("should not raise error: %v", err)
	}
	defer os.Remove(file.Name())

	expect := &Config{
		Apibase: "",
		Apikey:  "DUMMY APIKEY",
		Root:    "/tmp/mackerel-container-agent",
	}

	confLoader := NewLoader(file.Name(), 300*time.Millisecond)
	conf, err := confLoader.Load()
	if err != nil {
		t.Errorf("should not raise error: %v", err)
	}
	if !reflect.DeepEqual(conf, expect) {
		t.Errorf("expect %#v, got %#v", expect, conf)
	}

	ctx, cancel := context.WithCancel(context.Background())
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

	go func() {
		time.Sleep(800 * time.Millisecond)
		ioutil.WriteFile(file.Name(), []byte(`
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
	conf, err = confLoader.Load()
	if err != nil {
		t.Errorf("should not raise error: %v", err)
	}
	if !reflect.DeepEqual(conf, expect2) {
		t.Errorf("expect %#v, got %#v", expect2, conf)
	}
}

func TestLoaderStartCancel(t *testing.T) {
	file, err := newConfigFile(`
apikey: 'DUMMY APIKEY'
root: '/tmp/mackerel-container-agent'
`)
	if err != nil {
		t.Errorf("should not raise error: %v", err)
	}
	defer os.Remove(file.Name())

	expect := &Config{
		Apibase: "",
		Apikey:  "DUMMY APIKEY",
		Root:    "/tmp/mackerel-container-agent",
	}

	confLoader := NewLoader(file.Name(), 300*time.Millisecond)
	conf, err := confLoader.Load()
	if err != nil {
		t.Errorf("should not raise error: %v", err)
	}
	if !reflect.DeepEqual(conf, expect) {
		t.Errorf("expect %#v, got %#v", expect, conf)
	}

	ctx, cancel := context.WithCancel(context.Background())
	confCh := confLoader.Start(ctx)

	go func() {
		time.Sleep(500 * time.Millisecond)
		cancel()
	}()

	<-confCh // when the context is done, loader should stop the polling loop
	<-ctx.Done()
}
