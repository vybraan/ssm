package sshconf_test

import (
	"fmt"
	"log"
	"testing"

	"github.com/lfaoro/ssm/pkg/sshconf"
)

func TestParse(t *testing.T) {
	cfg := sshconf.New()
	err := cfg.ParsePath("../../data/config_example")
	if err != nil {
		log.Println(err)
		t.FailNow()
	}

	for _, h := range cfg.Hosts {
		fmt.Println(h)
		fmt.Println(h.Name)
		for i, k := range h.Options.Keys() {
			fmt.Println(k, h.Options.Values()[i])
		}
	}
	err = cfg.ParsePath("./nonexistent")
	if err == nil {
		t.FailNow()
	}
}
