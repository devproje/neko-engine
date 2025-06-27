package main

import (
	"fmt"
	"os"

	"github.com/devproje/commando"
	"github.com/devproje/commando/types"
	"github.com/devproje/neko-engine/common"
	"github.com/devproje/neko-engine/config"
	"github.com/devproje/neko-engine/internal"
)

var (
	version   = "0.1.0"
	branch    = "unknown"
	hash      = "unknown"
	buildTime = "unknown"
	goVersion = "unknown"
	channel   = "dev"
)

func init() {
	config.Version = version
	config.Branch = branch
	config.Hash = hash
	config.BuildTime = buildTime
	config.GoVersion = goVersion
	config.Channel = channel
}

func checkVersion(n *commando.Node) error {
	showVersion()
	return nil
}

func serve(n *commando.Node) error {
	sl := common.New()
	internal.NewInternalServer(sl)

	return nil
}

func showVersion() {
	fmt.Printf("neko-engine %s\n", version)
	fmt.Printf("Channel: %s\n", channel)
	fmt.Printf("Branch: %s\n", branch)
	fmt.Printf("Commit: %s\n", hash)
	fmt.Printf("Build Time: %s\n", buildTime)
	fmt.Printf("Go Version: %s\n", goVersion)
}

func main() {
	command := commando.NewCommando(os.Args[1:])

	command.Root("serve", "running neko-engine backend service", serve,
		types.OptionData{
			Name:  "debug",
			Desc:  "debugging gin",
			Short: []string{"d"},
			Type:  types.BOOLEAN,
		},
	)

	command.Root("version", "checking neko-engine version", checkVersion)

	if err := command.Execute(); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%v\n", err)
		return
	}
}
