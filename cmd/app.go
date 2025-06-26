package main

import (
	"os"

	"github.com/devproje/commando"
	"github.com/devproje/commando/types"
)

func serve(n *commando.Node) error {
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
}
