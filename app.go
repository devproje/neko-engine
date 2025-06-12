package main

import (
	"fmt"
	"os"
	"time"

	"github.com/devproje/commando"
	"github.com/devproje/commando/option"
	"github.com/devproje/commando/types"
	"github.com/devproje/neko-engine/config"
	"github.com/devproje/neko-engine/routes"
	"github.com/devproje/neko-engine/service"
	"github.com/gin-gonic/gin"
)

var (
	version string
	branch  string
	hash    string
)

func serve(n *commando.Node) error {
	cnf := config.Load()
	fmt.Printf("Neko Engine %s-%s (%s)\n", version, branch, hash)
	debug, _ := option.ParseBool(*n.MustGetOpt("debug"), n)
	gin.SetMode(gin.ReleaseMode)
	config.Debug = false
	if debug {
		config.Debug = true
		gin.SetMode(gin.DebugMode)
	}

	app := gin.Default()
	routes.New(app)

	if err := app.Run(fmt.Sprintf("%s:%d", cnf.Server.Host, cnf.Server.Port)); err != nil {
		return err
	}

	return nil
}

func versionCheck(n *commando.Node) error {
	fmt.Printf("neko-engine v%s %s(%s)\n", version, branch, hash)

	return nil
}

func reset() {
	acc := service.NewAccountService()

	for {
		now := time.Now()

		if now.Hour() == 0 && now.Minute() == 0 && now.Second() == 0 {
			acc.ResetCount()
			fmt.Printf("[INFO] daily prompt count is resetting\n")
		}

		time.Sleep(time.Second * 1)
	}
}

func main() {
	command := commando.NewCommando(os.Args[1:])
	go reset()

	command.Root("serve", "running bot server", serve, types.OptionData{
		Name:  "debug",
		Desc:  "serve gin debugging mode",
		Short: []string{"d"},
		Type:  types.BOOLEAN,
	})

	command.Root("version", "checking engine version", versionCheck)

	if err := command.Execute(); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}
