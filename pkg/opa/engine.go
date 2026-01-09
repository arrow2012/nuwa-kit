package opa

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/arrow2012/nuwa-kit/pkg/log"
	"github.com/open-policy-agent/opa/rego"
	"go.uber.org/zap"
)

// Engine defines the policy evaluation interface
type Engine interface {
	Evaluate(ctx context.Context, input interface{}) (bool, map[string]interface{}, error)
	Close()
}

// RegoEngine implements Engine using OPA Rego
type RegoEngine struct {
	mutex       sync.RWMutex
	query       rego.PreparedEvalQuery
	policyPath  string
	queryString string
	stopChan    chan struct{}
}

// NewRegoEngine creates a new engine and loads policy from file
// query: e.g. "data.iam.decision"
func NewRegoEngine(ctx context.Context, policyPath string, query string) (*RegoEngine, error) {
	e := &RegoEngine{
		policyPath:  policyPath,
		queryString: query,
		stopChan:    make(chan struct{}),
	}

	if err := e.loadPolicy(ctx); err != nil {
		return nil, err
	}

	// Watch for changes (Simple Polling for MVP)
	go e.watchPolicy(ctx)

	return e, nil
}

func (e *RegoEngine) loadPolicy(ctx context.Context) error {
	content, err := os.ReadFile(e.policyPath)
	if err != nil {
		return fmt.Errorf("failed to read policy file: %w", err)
	}

	fileName := filepath.Base(e.policyPath)

	// Create Rego Object
	r := rego.New(
		rego.Query(e.queryString),
		rego.Module(fileName, string(content)),
	)

	// Prepare for Evaluation (Optimization)
	query, err := r.PrepareForEval(ctx)
	if err != nil {
		return fmt.Errorf("failed to prepare rego: %w", err)
	}

	e.mutex.Lock()
	e.query = query
	e.mutex.Unlock()

	log.Info("OPA Policy loaded successfully", zap.String("path", e.policyPath))
	return nil
}

func (e *RegoEngine) watchPolicy(ctx context.Context) {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	var lastModTime time.Time

	if info, err := os.Stat(e.policyPath); err == nil {
		lastModTime = info.ModTime()
	}

	for {
		select {
		case <-e.stopChan:
			return
		case <-ticker.C:
			info, err := os.Stat(e.policyPath)
			if err != nil {
				continue
			}
			if info.ModTime().After(lastModTime) {
				log.Info("Policy file changed, reloading...", zap.String("path", e.policyPath))
				if err := e.loadPolicy(ctx); err == nil {
					lastModTime = info.ModTime()
				} else {
					log.Error("Failed to reload policy", zap.Error(err))
				}
			}
		}
	}
}

func (e *RegoEngine) Evaluate(ctx context.Context, input interface{}) (bool, map[string]interface{}, error) {
	e.mutex.RLock()
	query := e.query
	e.mutex.RUnlock()

	rs, err := query.Eval(ctx, rego.EvalInput(input))
	if err != nil {
		return false, nil, err
	}

	if len(rs) == 0 {
		return false, nil, nil
	}

	// Expecting object: { "allowed": bool, "reasons": ... }
	// This structure depends on the rego policy output format.
	// Since this is now a generic engine, strictly expecting "allowed" and "reasons" maps might be too specific?
	// But the user asked to just parameterize the query, not necessarily the output format.
	// I'll keep it as is but add a comment. For true generic, we might return interface{}.
	// But let's assume standard Nuwa policy format.

	if len(rs[0].Expressions) == 0 {
		return false, nil, fmt.Errorf("no result expressions")
	}

	resultMap, ok := rs[0].Expressions[0].Value.(map[string]interface{})
	if !ok {
		// If the query returns a boolean directly (e.g. data.iam.allow)
		if allowed, ok := rs[0].Expressions[0].Value.(bool); ok {
			return allowed, nil, nil
		}
		return false, nil, fmt.Errorf("unexpected opa result type: %T", rs[0].Expressions[0].Value)
	}

	allowed, _ := resultMap["allowed"].(bool)
	reasons, _ := resultMap["reasons"].(map[string]interface{})

	return allowed, reasons, nil
}

func (e *RegoEngine) Close() {
	close(e.stopChan)
}
