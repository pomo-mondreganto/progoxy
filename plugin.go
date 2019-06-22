package main

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"plugin"
)

type Processor interface {
	Init(map[string]interface{}, ...interface{}) error
	Process(buf []byte) []byte
}

type PluginWrapper struct {
	proc Processor
}

func (pw *PluginWrapper) Load(pluginName string, pluginConfig map[string]interface{}) {
	filepath := fmt.Sprintf("%s/plugins/%s.so", viper.Get("resources_path"), pluginName)
	p, err := plugin.Open(filepath)
	if err != nil {
		logrus.Fatalf("Error opening plugin %s: %v", pluginName, err)
	}
	procGetterInter, err := p.Lookup("GetProcessor")
	if err != nil {
		logrus.Fatalf("Error loading processor getter for plugin %s: %v", pluginName, err)
	}

	procGetter, ok := procGetterInter.(func() interface{})
	if !ok {
		logrus.Fatalf("Error casting processor getter for plugin %s: %T", pluginName, procGetterInter)
	}

	procInter := procGetter()

	proc, ok := procInter.(Processor)
	if !ok {
		logrus.Fatalf("Error casting processor interface for plugin %s: %T", pluginName, procInter)
	}

	pw.proc = proc

	err = pw.proc.Init(pluginConfig, viper.GetString("resources_path"))
	if err != nil {
		logrus.Fatalf("Error initializing plugin %s: %v", pluginName, err)
	}
}

func (pw *PluginWrapper) Run(input []byte) []byte {
	return pw.proc.Process(input)
}
