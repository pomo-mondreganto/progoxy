package main

import (
	"fmt"
	"github.com/sirupsen/logrus"
)

type Service struct {
	Name       string
	SrcAddr    string
	DstAddr    string
	SrcPlugins []*PluginWrapper
	DstPlugins []*PluginWrapper
}

func loadPlugins(pluginsMap map[string]interface{}) []*PluginWrapper {
	wrappers := make([]*PluginWrapper, 0, len(pluginsMap))

	for pluginName, pluginConfig := range pluginsMap {
		pluginConfigMap := pluginConfig.(map[string]interface{})
		pw := &PluginWrapper{}
		pw.Load(pluginName, pluginConfigMap)
		wrappers = append(wrappers, pw)
	}

	logrus.Infof("Successfully loaded %d plugins", len(wrappers))

	return wrappers
}

func (sv *Service) Load(name string, serviceMap map[string]interface{}) {
	srcConfig, ok := serviceMap["source"].(map[string]interface{})
	if !ok {
		logrus.Fatal("No source config provided for service: ", serviceMap)
	}

	dstConfig, ok := serviceMap["destination"].(map[string]interface{})
	if !ok {
		logrus.Fatal("No destination config provided for service: ", serviceMap)
	}

	srcPort, ok := srcConfig["port"].(int)
	if !ok {
		logrus.Fatal("No source port provided for service: ", serviceMap)
	}

	dstPort, ok := dstConfig["port"].(int)
	if !ok {
		logrus.Fatal("No destination port provided for service: ", serviceMap)
	}

	srcHost, ok := srcConfig["host"].(string)
	if !ok {
		srcHost = "0.0.0.0"
	}

	dstHost, ok := dstConfig["host"].(string)
	if !ok {
		dstHost = "127.0.0.1"
	}

	var srcPlugins []*PluginWrapper
	srcPluginsMap, ok := srcConfig["plugins"].(map[string]interface{})
	if ok {
		srcPlugins = loadPlugins(srcPluginsMap)
	} else {
		srcPlugins = make([]*PluginWrapper, 0)
	}

	var dstPlugins []*PluginWrapper
	dstPluginsMap, ok := dstConfig["plugins"].(map[string]interface{})
	if ok {
		dstPlugins = loadPlugins(dstPluginsMap)
	} else {
		dstPlugins = make([]*PluginWrapper, 0)
	}

	sv.Name = name
	sv.SrcAddr = fmt.Sprintf("%s:%d", srcHost, srcPort)
	sv.DstAddr = fmt.Sprintf("%s:%d", dstHost, dstPort)
	sv.SrcPlugins = srcPlugins
	sv.DstPlugins = dstPlugins
}
