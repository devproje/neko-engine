package main

import (
	"fmt"
	"os"

	"github.com/devproje/commando"
	"github.com/devproje/commando/types"
	"github.com/devproje/neko-engine/common"
	"github.com/devproje/neko-engine/internal"
)

func serve(n *commando.Node) error {
	sl := common.New()

	internal.NewInternalServer(sl)
	return nil
}

func main() {
	command := commando.NewCommando(os.Args[1:])

	command.Root("serve", "running neko-engine backend service", serve, types.OptionData{
		Name:  "debug",
		Desc:  "debbuging gin",
		Short: []string{"d"},
		Type:  types.BOOLEAN,
	})

	if err := command.Execute(); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%v\n", err)
		return
	}
}
