package vm

// Function represents a compiled function
type Function struct {
	Name         string        // function name
	Instructions []Instruction // bytecode instructions
	Constants    []Value       // constant pool
	NumParams    int           // number of parameters
	NumLocals    int           // number of local variables
	NumUpvalues  int           // number of upvalues
	IsVariadic   bool          // whether function accepts variable arguments
	SourceFile   string        // source file name
	LineNumbers  []int         // line number for each instruction
}

// NewFunction creates a new function
func NewFunction(name string) *Function {
	return &Function{
		Name:         name,
		Instructions: make([]Instruction, 0),
		Constants:    make([]Value, 0),
		NumParams:    0,
		NumLocals:    0,
		NumUpvalues:  0,
		IsVariadic:   false,
		SourceFile:   "",
		LineNumbers:  make([]int, 0),
	}
}

// AddInstruction adds an instruction to the function
func (f *Function) AddInstruction(inst Instruction, line int) int {
	f.Instructions = append(f.Instructions, inst)
	f.LineNumbers = append(f.LineNumbers, line)
	return len(f.Instructions) - 1
}

// AddConstant adds a constant to the function's constant pool
func (f *Function) AddConstant(value Value) int {
	// Check if constant already exists
	for i, constant := range f.Constants {
		if constant.Equals(value) {
			return i
		}
	}
	
	f.Constants = append(f.Constants, value)
	return len(f.Constants) - 1
}

// GetConstant returns the constant at the given index
func (f *Function) GetConstant(index int) (Value, bool) {
	if index < 0 || index >= len(f.Constants) {
		return NilValue, false
	}
	return f.Constants[index], true
}

// GetInstruction returns the instruction at the given index
func (f *Function) GetInstruction(index int) (Instruction, bool) {
	if index < 0 || index >= len(f.Instructions) {
		return 0, false
	}
	return f.Instructions[index], true
}

// GetLineNumber returns the line number for the instruction at the given index
func (f *Function) GetLineNumber(index int) int {
	if index < 0 || index >= len(f.LineNumbers) {
		return 0
	}
	return f.LineNumbers[index]
}

// NativeFunctionType represents the signature of a native function
type NativeFunctionType func(vm *VM, args []Value) (Value, error)

// NativeFunction represents a native (Go) function
type NativeFunction struct {
	Name     string             // function name
	Function NativeFunctionType // the actual function
	MinArgs  int                // minimum number of arguments
	MaxArgs  int                // maximum number of arguments (-1 for unlimited)
}

// NewNativeFunction creates a new native function
func NewNativeFunction(name string, fn NativeFunctionType, minArgs, maxArgs int) *NativeFunction {
	return &NativeFunction{
		Name:     name,
		Function: fn,
		MinArgs:  minArgs,
		MaxArgs:  maxArgs,
	}
}

// Call calls the native function
func (nf *NativeFunction) Call(vm *VM, args []Value) (Value, error) {
	// Check argument count
	if len(args) < nf.MinArgs {
		return NilValue, NewRuntimeError("function '%s' expects at least %d arguments, got %d", 
			nf.Name, nf.MinArgs, len(args))
	}
	
	if nf.MaxArgs >= 0 && len(args) > nf.MaxArgs {
		return NilValue, NewRuntimeError("function '%s' expects at most %d arguments, got %d", 
			nf.Name, nf.MaxArgs, len(args))
	}
	
	return nf.Function(vm, args)
}

// Upvalue represents an upvalue (captured variable)
type Upvalue struct {
	Location *Value // pointer to the variable location
	Closed   Value  // closed value (when variable goes out of scope)
	IsClosed bool   // whether the upvalue is closed
}

// NewUpvalue creates a new upvalue
func NewUpvalue(location *Value) *Upvalue {
	return &Upvalue{
		Location: location,
		Closed:   NilValue,
		IsClosed: false,
	}
}

// Get returns the upvalue's current value
func (uv *Upvalue) Get() Value {
	if uv.IsClosed {
		return uv.Closed
	}
	return *uv.Location
}

// Set sets the upvalue's value
func (uv *Upvalue) Set(value Value) {
	if uv.IsClosed {
		uv.Closed = value
	} else {
		*uv.Location = value
	}
}

// Close closes the upvalue
func (uv *Upvalue) Close() {
	if !uv.IsClosed {
		uv.Closed = *uv.Location
		uv.IsClosed = true
		uv.Location = nil
	}
}

// Closure represents a function closure
type Closure struct {
	Function *Function  // the function
	Upvalues []*Upvalue // captured upvalues
}

// NewClosure creates a new closure
func NewClosure(function *Function) *Closure {
	return &Closure{
		Function: function,
		Upvalues: make([]*Upvalue, function.NumUpvalues),
	}
}

// GetUpvalue returns the upvalue at the given index
func (c *Closure) GetUpvalue(index int) (*Upvalue, bool) {
	if index < 0 || index >= len(c.Upvalues) {
		return nil, false
	}
	return c.Upvalues[index], true
}

// SetUpvalue sets the upvalue at the given index
func (c *Closure) SetUpvalue(index int, upvalue *Upvalue) bool {
	if index < 0 || index >= len(c.Upvalues) {
		return false
	}
	c.Upvalues[index] = upvalue
	return true
}