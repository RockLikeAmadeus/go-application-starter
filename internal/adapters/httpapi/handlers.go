// Package httpapi is the driving HTTP adapter. It implements the strict
// server interface generated from api/openapi.yaml by delegating to the core's
// driving ports.
package httpapi

import (
	"github.com/RockLikeAmadeus/go-application-starter/internal/adapters/httpapi/openapi"
	"github.com/RockLikeAmadeus/go-application-starter/internal/core/ports"
)

// Handlers translates between HTTP DTOs and the core. Errors returned from
// these methods are rendered as problem documents by writeProblem.
type Handlers struct {
	// sample:start
	items ports.ItemService
	// sample:end
	entitlements ports.EntitlementsService
	isDataLocal  bool
	isSignInOn   bool
	warnings     []string
	clientSignIn ClientSignIn
}

var _ openapi.StrictServerInterface = (*Handlers)(nil)

// NewHandlers serves the driving ports in opts.
func NewHandlers(opts Options) *Handlers {
	return &Handlers{
		// sample:start
		items: opts.Items,
		// sample:end
		entitlements: opts.Entitlements,
		isDataLocal:  opts.IsDataLocal,
		isSignInOn:   opts.SignIn != nil,
		warnings:     opts.Warnings,
		clientSignIn: opts.ClientSignIn,
	}
}
