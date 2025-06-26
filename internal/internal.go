package internal

import (
	"fmt"
	"os"

	"github.com/devproje/neko-engine/common"
	"github.com/devproje/neko-engine/internal/route"
	"github.com/gin-gonic/gin"
)

func NewInternalServer(sl *common.ServiceLoader) {
	app := gin.Default()
	route.InternalRouter(app, sl)

	if err := app.Run(":8081"); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%v\n", err)
		return
	}
}
