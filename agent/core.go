package agent

import (
	"context"
	"time"

	"nofx/kernel"
)

// AgentType defines the role of an agent
type AgentType string

const (
	AgentTypeRiskManager AgentType = "risk_manager"
	AgentTypeAlphaHunter AgentType = "alpha_hunter"
	AgentTypeAnalyst     AgentType = "analyst"
	AgentTypeExecutor    AgentType = "executor"
)

// AgentStatus defines the current state of an agent
type AgentStatus string

const (
	AgentStatusIdle    AgentStatus = "idle"
	AgentStatusRunning AgentStatus = "running"
	AgentStatusError   AgentStatus = "error"
)

// AgentResult represents the output of an agent's analysis
type AgentResult struct {
	AgentID     string      `json:"agent_id"`
	Type        AgentType   `json:"type"`
	Timestamp   time.Time   `json:"timestamp"`
	Decision    string      `json:"decision"` // e.g. "approve", "reject", "buy", "sell"
	Confidence  float64     `json:"confidence"`
	Reasoning   string      `json:"reasoning"`
	Metadata    interface{} `json:"metadata,omitempty"`
}

// Agent defines the interface for all agents
type Agent interface {
	// ID returns the unique identifier of the agent
	ID() string
	
	// Type returns the role type of the agent
	Type() AgentType
	
	// Start starts the agent's background processes
	Start(ctx context.Context) error
	
	// Stop stops the agent
	Stop() error
	
	// Analyze performs analysis based on market context
	Analyze(ctx context.Context, marketCtx *kernel.Context) (*AgentResult, error)
	
	// Status returns the current status
	Status() AgentStatus
}

// BaseAgent provides common functionality for agents
type BaseAgent struct {
	id     string
	role   AgentType
	status AgentStatus
}

func NewBaseAgent(id string, role AgentType) *BaseAgent {
	return &BaseAgent{
		id:     id,
		role:   role,
		status: AgentStatusIdle,
	}
}

func (a *BaseAgent) ID() string {
	return a.id
}

func (a *BaseAgent) Type() AgentType {
	return a.role
}

func (a *BaseAgent) Status() AgentStatus {
	return a.status
}
