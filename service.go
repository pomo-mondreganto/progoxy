package main

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"io"
	"net"
	"os/exec"
)

type socketConnector struct {
	addr string
}

func (sc socketConnector) GetConnection() (io.ReadWriteCloser, error) {
	return net.Dial("tcp", sc.addr)
}

type commandConnector struct {
	command string
}

func (ca commandConnector) GetConnection() (io.ReadWriteCloser, error) {
	w, r := io.Pipe()

	crwc := CommandReadWriteCloserImpl{}
	crwc.ReadCloser = w

	cmd := exec.Command("sh", "-c", ca.command)
	cmd.Stdout = r
	cmd.Stderr = r

	crwc.WriteCloser, _ = cmd.StdinPipe()
	err = cmd.Start()

	return crwc, err
}

type CommandReadWriteCloserImpl struct {
	io.ReadCloser
	io.WriteCloser
}

func (c CommandReadWriteCloserImpl) Close() (err error) {
	err = c.ReadCloser.Close()
	if err != nil {
		_ = c.WriteCloser.Close()
		return
	}
	err = c.WriteCloser.Close()
	return
}

type Connector interface {
	GetConnection() (io.ReadWriteCloser, error)
}

type Service struct {
	Name       string
	SrcAddr    string
	SrcPlugins []*PluginWrapper
	DstPlugins []*PluginWrapper
	Connector  Connector
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
	serviceType, ok := serviceMap["type"].(string)
	if !ok {
		logrus.Fatal("No type config provided for service: ", serviceMap)
	}

	srcConfig, ok := serviceMap["source"].(map[string]interface{})
	if !ok {
		logrus.Fatal("No source config provided for service: ", serviceMap)
	}

	dstConfig, ok := serviceMap["destination"].(map[string]interface{})
	if !ok {
		logrus.Fatal("No destination config provided for service: ", serviceMap)
	}

	srcHost, ok := srcConfig["host"].(string)
	if !ok {
		srcHost = "0.0.0.0"
	}

	srcPort, ok := srcConfig["port"].(int)
	if !ok {
		logrus.Fatal("No source port provided for service: ", serviceMap)
	}

	if serviceType == "socket" {
		dstHost, ok := dstConfig["host"].(string)
		if !ok {
			dstHost = "127.0.0.1"
		}

		dstPort, ok := dstConfig["port"].(int)
		if !ok {
			logrus.Fatal("No destination port provided for service: ", serviceMap)
		}

		sv.Connector = socketConnector{fmt.Sprintf("%s:%d", dstHost, dstPort)}
	} else if serviceType == "command" {
		command, ok := dstConfig["command"].(string)
		if !ok {
			logrus.Fatal("No command provided for service: ", serviceMap)
		}
		sv.Connector = commandConnector{command}
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
	sv.SrcPlugins = srcPlugins
	sv.DstPlugins = dstPlugins
}
