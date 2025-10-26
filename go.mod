module github.com/linearbits/erp-pos-module

go 1.21

require (
	github.com/go-chi/chi/v5 v5.0.10
	github.com/jmoiron/sqlx v1.3.5
	github.com/linearbits/erp-backend/pkg/module-sdk v0.0.0
	go.uber.org/zap v1.26.0
)

require go.uber.org/multierr v1.10.0 // indirect

// Use local SDK
replace github.com/linearbits/erp-backend/pkg/module-sdk => /Users/ahmedsamir/ai-workspace/linearbits-erp/erp/backend/pkg/module-sdk
