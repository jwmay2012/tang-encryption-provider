package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/flatheadmill/tang-encryption-provider/logger"
	"github.com/flatheadmill/tang-encryption-provider/plugin"
	"github.com/kelseyhightower/envconfig"
	"github.com/lainio/err2/try"
)

type Specification struct {
	ServerUrl  string `envconfig:"server_url"`
	Thumbprint string
	UnixSocket string `envconfig:"unix_socket" default:"/var/run/kmsplugin/socket.sock"`
}

func main() {
	var spec Specification
	try.To(envconfig.Process("tang_kms", &spec))
	log := logger.New(os.Stdout)
	log.Console()

	log.MsgWithFields(map[string]interface{}{"thumbprint": spec.Thumbprint, "unix_socket": spec.UnixSocket}, "")

	err := run(try.To1(plugin.New(spec.ServerUrl, spec.Thumbprint, spec.UnixSocket, log)))

	if err != nil {
		fmt.Printf("exited with error: %T %v\n", err, err)
	}
}

func run(plug *plugin.Plugin) error {
	signalsChannel := make(chan os.Signal, 1)
	signal.Notify(signalsChannel, syscall.SIGINT, syscall.SIGTERM)

	rpc, rpcErrorChannel := plug.ServeKMSRequests()
	if rpc != nil {
		defer rpc.GracefulStop()
	}

	select {
	case sig := <-signalsChannel:
		fmt.Printf("captured %v, shutting down kms-plugin\n", sig)
		return nil
	case err := <-rpcErrorChannel:
		return err
	}
}