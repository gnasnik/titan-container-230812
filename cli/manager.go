package cli

import (
	"github.com/urfave/cli/v2"
)

// ManagerCMDs manager cmd
var ManagerCMDs = []*cli.Command{
	WithCategory("providers", providersCmds),
}
