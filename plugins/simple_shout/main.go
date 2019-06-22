package main

import (
	"github.com/sirupsen/logrus"
	"strings"
)

type Processor struct{}

func (p *Processor) Init(config map[string]interface{}, _ ...interface{}) error {
	return nil
}

func (p *Processor) Process(buf []byte) []byte {
	logrus.Debugf("Shouter got data: %s", buf)
	return []byte(strings.ToUpper(string(buf)))
}

func GetProcessor() interface{} {
	return &Processor{}
}
