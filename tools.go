// +build tools

package tools

import (
	_ "github.com/go-playground/overalls"
	_ "github.com/mgechev/revive"
	_ "github.com/pingcap/failpoint/failpoint-ctl"
	_ "github.com/sasha-s/go-deadlock"
	_ "golang.org/x/tools/cmd/goimports"
)
