package SSAruntime

// Mode is a bitmask of options influencing the interpreter.
type Mode uint

// Mode is a bitmask of options influencing the tracing.
type TraceMode uint

const (
	// Disable recover() in target programs; show interpreter crash instead.
	DisableRecover Mode = 1 << iota
)

const (
	// Print a trace of all instructions as they are interpreted.
	EnableTracing  TraceMode = 1 << iota

	// Print higher-level statement boundary tracing
	EnableStmtTracing
)

type Status int

const (
	StRunning Status = iota
	StComplete
	StPanic
)
