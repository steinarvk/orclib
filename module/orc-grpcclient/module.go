package orcgrpcclient

import (
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/steinarvk/orc"
	"google.golang.org/grpc"

	orcgrpcclientcommon "github.com/steinarvk/orclib/module/orc-grpcclientcommon"
)

type Module struct {
	name string
	Conn *grpc.ClientConn
}

func New(name string) *Module {
	return &Module{
		name: name,
	}
}

func (m *Module) ModuleName() string { return fmt.Sprintf("gRPCServer(%s)", m.name) }

func (m *Module) OnRegister(hooks orc.ModuleHooks) {
	var serverAddr string

	hooks.OnUse(func(u orc.UseContext) {
		u.Use(orcgrpcclientcommon.M)
		u.Flags.StringVar(&serverAddr, m.name+"_server", "", "gRPC server to connect to for "+m.name)
	})

	hooks.OnStart(func() error {
		logrus.Infof("Dialing gRPC connection to %s (%q)", m.name, serverAddr)
		conn, err := grpc.Dial(serverAddr, orcgrpcclientcommon.M.DialOptions...)
		if err != nil {
			return fmt.Errorf("Failed to dial %s (%q): %v", m.name, serverAddr, err)
		}
		m.Conn = conn
		return nil
	})

	hooks.OnStop(func() error {
		logrus.Infof("Closing gRPC connection to %s", m.name)
		m.Conn.Close()
		return nil
	})
}
