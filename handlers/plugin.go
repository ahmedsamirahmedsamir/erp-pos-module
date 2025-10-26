package main

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/jmoiron/sqlx"
	sdk "github.com/linearbits/erp-backend/pkg/module-sdk"
	"go.uber.org/zap"
)

// POSPlugin implements the ModulePlugin interface
type POSPlugin struct {
	db      *sqlx.DB
	logger  *zap.Logger
	handler *POSHandler
}

// NewPOSPlugin creates a new plugin instance
func NewPOSPlugin() sdk.ModulePlugin {
	return &POSPlugin{}
}

// Initialize initializes the plugin
func (p *POSPlugin) Initialize(db *sqlx.DB, logger *zap.Logger) error {
	p.db = db
	p.logger = logger
	p.handler = NewPOSHandler(db, logger)
	p.logger.Info("POS module initialized")
	return nil
}

// GetModuleCode returns the module code
func (p *POSPlugin) GetModuleCode() string {
	return "pos"
}

// GetModuleVersion returns the module version
func (p *POSPlugin) GetModuleVersion() string {
	return "1.0.0"
}

// Cleanup performs cleanup
func (p *POSPlugin) Cleanup() error {
	p.logger.Info("Cleaning up POS module")
	return nil
}

// GetHandler returns a handler function for a given route and method
func (p *POSPlugin) GetHandler(route string, method string) (http.HandlerFunc, error) {
	route = strings.TrimPrefix(route, "/")
	method = strings.ToUpper(method)

	handlers := map[string]http.HandlerFunc{
		"GET /transactions":  p.handler.GetTransactions,
		"POST /transactions": p.handler.CreateTransaction,
		"GET /shifts":        p.handler.GetShifts,
		"POST /shifts":       p.handler.CreateShift,
	}

	key := method + " " + route
	if handler, ok := handlers[key]; ok {
		return handler, nil
	}

	return nil, fmt.Errorf("handler not found for route: %s %s", method, route)
}

// Handler is the exported symbol
var Handler = NewPOSPlugin
