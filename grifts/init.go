package grifts

import (
	"github.com/FANIoT/pm/actions"
	"github.com/gobuffalo/buffalo"
)

func init() {
	buffalo.Grifts(actions.App())
}
