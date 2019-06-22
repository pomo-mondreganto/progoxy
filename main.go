package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"os"
	"os/signal"
	"time"
)

func init() {
	resourcesDir := flag.String("resources", "./resources", "Directory with config and plugins")
	flag.Parse()

	configPath := fmt.Sprintf("%s/config.yml", *resourcesDir)
	viper.SetConfigFile(configPath)

	viper.Set("resources_path", *resourcesDir)

	viper.SetDefault("log_everything", true)
	viper.SetDefault("idle_timeout", 10*time.Second)
	viper.SetDefault("max_read_bytes", 1<<16)

	err := viper.ReadInConfig()
	if err != nil {
		logrus.Fatal("Error while reading config: ", err)
	}

	if viper.GetBool("log_everything") {
		logrus.SetLevel(logrus.DebugLevel)
	} else {
		logrus.SetLevel(logrus.InfoLevel)
	}
}

func loadServices() []*Service {
	servicesMap := viper.GetStringMap("services")

	services := make([]*Service, 0, len(servicesMap))
	for serviceName, serviceMap := range servicesMap {
		service := &Service{}
		service.Load(serviceName, serviceMap.(map[string]interface{}))
		services = append(services, service)
	}
	logrus.Infof("Successfully loaded %d services", len(services))

	return services
}

func startServers(services []*Service) []*ServiceServer {
	servers := make([]*ServiceServer, 0, len(services))
	for _, service := range services {
		server := &ServiceServer{
			Srv:          service,
			IdleTimeout:  viper.GetDuration("idle_timeout"),
			MaxReadBytes: viper.GetInt64("max_read_bytes"),
		}
		server.Init()
		go server.Serve()
		servers = append(servers, server)
	}
	return servers
}

func main() {
	services := loadServices()
	servers := startServers(services)

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	signal.Notify(c, os.Kill)

	<-c

	logrus.Info("Shutting down servers")
	for _, server := range servers {
		ctx, cancel := context.WithTimeout(context.Background(), 2*server.IdleTimeout)
		err = server.Shutdown(ctx)
		if err != nil {
			logrus.Fatal("Error shutting down server: ", err)
		}
		cancel()
	}
}
