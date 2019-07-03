package config

import (
	"os"
	"reflect"
	"testing"
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
}
