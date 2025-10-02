package debug

import (
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/xingleixu/TG-Script/ast"
	"github.com/xingleixu/TG-Script/lexer"
	"github.com/xingleixu/TG-Script/vm"
)

// DebugLevel represents the level of debugging output
type DebugLevel int

const (
	DebugOff DebugLevel = iota
	DebugError
	DebugWarn
	DebugInfo
	DebugVerbose
	DebugTrace
)

// String returns the string representation of the debug level
func (level DebugLevel) String() string {
	switch level {
	case DebugOff:
		return "OFF"
	case DebugError:
		return "ERROR"
	case DebugWarn:
		return "WARN"
	case DebugInfo:
		return "INFO"
	case DebugVerbose:
		return "VERBOSE"
	case DebugTrace:
		return "TRACE"
	default:
		return "UNKNOWN"
	}
}

// Debugger provides debugging functionality for the TG-Script VM
type Debugger struct {
	level       DebugLevel
	output      io.Writer
	useColors   bool
	timestamps  bool
	breakpoints map[int]bool
	stepMode    bool
	stepCount   int
	maxSteps    int
}

// NewDebugger creates a new debugger instance
func NewDebugger() *Debugger {
	return &Debugger{
		level:       DebugInfo,
		output:      os.Stderr,
		useColors:   true,
		timestamps:  true,
		breakpoints: make(map[int]bool),
		stepMode:    false,
		stepCount:   0,
		maxSteps:    1000,
	}
}

// SetLevel sets the debug level
func (d *Debugger) SetLevel(level DebugLevel) {
	d.level = level
}

// SetOutput sets the output writer
func (d *Debugger) SetOutput(w io.Writer) {
	d.output = w
}

// SetColors enables or disables colored output
func (d *Debugger) SetColors(enabled bool) {
	d.useColors = enabled
}

// SetTimestamps enables or disables timestamps
func (d *Debugger) SetTimestamps(enabled bool) {
	d.timestamps = enabled
}

// AddBreakpoint adds a breakpoint at the specified PC
func (d *Debugger) AddBreakpoint(pc int) {
	d.breakpoints[pc] = true
}

// RemoveBreakpoint removes a breakpoint at the specified PC
func (d *Debugger) RemoveBreakpoint(pc int) {
	delete(d.breakpoints, pc)
}

// SetStepMode enables or disables step mode
func (d *Debugger) SetStepMode(enabled bool) {
	d.stepMode = enabled
	d.stepCount = 0
}

// SetMaxSteps sets the maximum number of steps in step mode
func (d *Debugger) SetMaxSteps(max int) {
	d.maxSteps = max
}

// logf formats and logs a message with the specified level and category
func (d *Debugger) logf(level DebugLevel, category, format string, args ...interface{}) {
	if level > d.level {
		return
	}

	var prefix string
	if d.timestamps {
		prefix = time.Now().Format("15:04:05.000")
	}

	if d.useColors {
		switch level {
		case DebugError:
			prefix += " \033[31m[ERROR]\033[0m"
		case DebugWarn:
			prefix += " \033[33m[WARN]\033[0m"
		case DebugInfo:
			prefix += " \033[32m[INFO]\033[0m"
		case DebugVerbose:
			prefix += " \033[36m[VERBOSE]\033[0m"
		case DebugTrace:
			prefix += " \033[37m[TRACE]\033[0m"
		}
	} else {
		prefix += fmt.Sprintf(" [%s]", level.String())
	}

	if category != "" {
		prefix += fmt.Sprintf(" [%s]", category)
	}

	message := fmt.Sprintf(format, args...)
	fmt.Fprintf(d.output, "%s %s\n", prefix, message)
}

// Error logs an error message
func (d *Debugger) Error(format string, args ...interface{}) {
	d.logf(DebugError, "", format, args...)
}

// Warn logs a warning message
func (d *Debugger) Warn(format string, args ...interface{}) {
	d.logf(DebugWarn, "", format, args...)
}

// Info logs an info message
func (d *Debugger) Info(format string, args ...interface{}) {
	d.logf(DebugInfo, "", format, args...)
}

// Verbose logs a verbose message
func (d *Debugger) Verbose(format string, args ...interface{}) {
	d.logf(DebugVerbose, "", format, args...)
}

// Trace logs a trace message
func (d *Debugger) Trace(format string, args ...interface{}) {
	d.logf(DebugTrace, "", format, args...)
}

// LogToken logs a token with its position
func (d *Debugger) LogToken(token lexer.Token, literal string, pos lexer.Position) {
	if d.level < DebugTrace {
		return
	}
	
	posStr := fmt.Sprintf("%d:%d", pos.Line, pos.Column)
	d.logf(DebugTrace, "TOKEN", "%-12s %-20s at %s", token.String(), literal, posStr)
}

// LogAST logs an AST node
func (d *Debugger) LogAST(node ast.Node, depth int) {
	if d.level < DebugTrace {
		return
	}
	
	indent := strings.Repeat("  ", depth)
	pos := node.Pos()
	posStr := fmt.Sprintf("%d:%d", pos.Line, pos.Column)
	
	// Handle different node types
	var nodeInfo string
	switch n := node.(type) {
	case *ast.Identifier:
		nodeInfo = fmt.Sprintf("Identifier: %s", n.Name)
	case *ast.IntegerLiteral:
		nodeInfo = fmt.Sprintf("IntegerLiteral: %d", n.Value)
	case *ast.FloatLiteral:
		nodeInfo = fmt.Sprintf("FloatLiteral: %f", n.Value)
	case *ast.StringLiteral:
		nodeInfo = fmt.Sprintf("StringLiteral: %s", n.Value)
	case *ast.BooleanLiteral:
		nodeInfo = fmt.Sprintf("BooleanLiteral: %t", n.Value)
	default:
		nodeInfo = fmt.Sprintf("%T: %s", node, node.String())
	}
	
	d.logf(DebugTrace, "AST", "%s%s at %s", indent, nodeInfo, posStr)
}

// LogVMState logs the current VM state
func (d *Debugger) LogVMState(machine *vm.VM) {
	if d.level < DebugVerbose {
		return
	}
	
	frameInfo := "no frame"
	if machine.CurrentFrame != nil {
		frameInfo = fmt.Sprintf("PC:%d", machine.CurrentFrame.PC)
	}
	
	d.logf(DebugVerbose, "VM", "Frame: %d, %s", machine.FrameIndex, frameInfo)
	
	// Log some register values
	for i := 0; i < 5; i++ {
		val := machine.Registers[i]
		if !val.IsNil() {
			d.logf(DebugVerbose, "VM", "  Reg[%d]: %s", i, val.ToString())
		}
	}
	
	// Log global variables
	if len(machine.Globals) > 0 {
		d.logf(DebugVerbose, "VM", "Globals: %d", len(machine.Globals))
		count := 0
		for name, value := range machine.Globals {
			if count >= 3 {
				d.logf(DebugVerbose, "VM", "  ... and %d more", len(machine.Globals)-3)
				break
			}
			d.logf(DebugVerbose, "VM", "  %s: %s", name, value.ToString())
			count++
		}
	}
}

// LogInstruction logs a VM instruction
func (d *Debugger) LogInstruction(inst vm.Instruction, pc int) {
	if d.level < DebugTrace {
		return
	}
	
	d.logf(DebugTrace, "INST", "PC:%04d %s", pc, inst.String())
	
	// Log opcode details
	opcode := inst.GetOpCode()
	info := vm.OpCodeInfos[opcode]
	d.logf(DebugTrace, "INST", "  OpCode: %s (format: %v)", info.Name, info.Format)
}

// LogError logs a detailed error with context
func (d *Debugger) LogError(err error, context string, pos lexer.Position) {
	posStr := fmt.Sprintf("%d:%d", pos.Line, pos.Column)
	d.logf(DebugError, "ERROR", "%s at %s: %v", context, posStr, err)
}

// CheckBreakpoint checks if execution should break at the given PC
func (d *Debugger) CheckBreakpoint(pc int) bool {
	if d.breakpoints[pc] {
		d.logf(DebugInfo, "BREAK", "Breakpoint hit at PC:%d", pc)
		return true
	}
	
	if d.stepMode {
		d.stepCount++
		if d.stepCount >= d.maxSteps {
			d.logf(DebugInfo, "STEP", "Step limit reached (%d steps)", d.maxSteps)
			return true
		}
	}
	
	return false
}

// LogCompilerPhase logs compiler phase information
func (d *Debugger) LogCompilerPhase(phase string, details string) {
	if d.level < DebugVerbose {
		return
	}
	d.logf(DebugVerbose, "COMPILER", "%s: %s", phase, details)
}

// LogParserState logs parser state information
func (d *Debugger) LogParserState(currentToken lexer.Token, peekToken lexer.Token, context string) {
	if d.level < DebugTrace {
		return
	}
	d.logf(DebugTrace, "PARSER", "%s - Current: %s, Peek: %s", context, currentToken.String(), peekToken.String())
}

// LogFunctionCall logs function call information
func (d *Debugger) LogFunctionCall(name string, args []vm.Value) {
	if d.level < DebugVerbose {
		return
	}
	
	argStrs := make([]string, len(args))
	for i, arg := range args {
		argStrs[i] = arg.ToString()
	}
	
	d.logf(DebugVerbose, "CALL", "%s(%s)", name, strings.Join(argStrs, ", "))
}

// LogReturn logs function return information
func (d *Debugger) LogReturn(value vm.Value) {
	if d.level < DebugVerbose {
		return
	}
	d.logf(DebugVerbose, "RETURN", "-> %s", value.ToString())
}

// Global debugger instance
var GlobalDebugger = NewDebugger()

// Convenience functions for global debugger
func SetDebugLevel(level DebugLevel) {
	GlobalDebugger.SetLevel(level)
}

func Error(format string, args ...interface{}) {
	GlobalDebugger.Error(format, args...)
}

func Warn(format string, args ...interface{}) {
	GlobalDebugger.Warn(format, args...)
}

func Info(format string, args ...interface{}) {
	GlobalDebugger.Info(format, args...)
}

func Verbose(format string, args ...interface{}) {
	GlobalDebugger.Verbose(format, args...)
}

func Trace(format string, args ...interface{}) {
	GlobalDebugger.Trace(format, args...)
}