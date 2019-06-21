package main

import (
	"errors"
	"github.com/sirupsen/logrus"
	"regexp"
)

var (
	err       error
	InvConfig = errors.New("invalid config")
)

type Processor struct {
	dropReg       *regexp.Regexp
}

func (p *Processor) Init(config map[string]interface{}) error {
	p.dropReg, err = regexp.Compile(config["regex"].(string))
	if err != nil {
		return InvConfig
	}

	logrus.Infof("Got regex %s", p.dropReg)
	return nil
}

func (p *Processor) Process(buf []byte) []byte {
	logrus.Debugf("Drop regex got: %s", buf)
	if p.dropReg.Match(buf) {
		return make([]byte, 0)
	}
	return buf
}

func GetProcessor() interface{} {
	return &Processor{}
}
