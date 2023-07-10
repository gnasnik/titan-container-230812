package types

import (
	"github.com/filecoin-project/go-jsonrpc/auth"
)

type OpenRPCDocument map[string]interface{}

type JWTPayload struct {
	Allow []auth.Permission
	ID    string
	//TODO remove NodeID later, any role id replace as ID
	NodeID string
	// Extend is json string
	Extend string
}
