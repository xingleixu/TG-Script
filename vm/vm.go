package vm

import (
	"fmt"
	"math"
)

// VM configuration constants
const (
	MaxRegisters    = 256  // maximum number of registers per frame
	MaxFrames       = 1024 // maximum call stack depth
	MaxGlobals      = 1024 // maximum number of global variables
	MaxConstants    = 1024 // maximum number of constants per function
	DefaultStackSize = 2048 // default stack size
)

// CallFrame represents a function call frame
type CallFrame struct {
	Closure     *Closure // function closure
	PC          int      // program counter
	BaseReg     int      // base register for this frame
	NumRegs     int      // number of registers used by this frame
	ReturnAddr  int      // return address (register to store result)
	NumResults  int      // number of expected return values
}

// VM represents the virtual machine
type VM struct {
	// Register file - the main execution context
	Registers [MaxRegisters]Value
	
	// Call stack
	Frames      [MaxFrames]CallFrame
	FrameIndex  int
	CurrentFrame *CallFrame
	
	// Global variables
	Globals map[string]Value
	
	// Native functions
	NativeFunctions map[string]*NativeFunction
	
	// Open upvalues (for closure capture)
	OpenUpvalues []*Upvalue
	
	// Execution state
	Running bool
	Error   error
	
	// Debug information
	DebugMode bool
	Breakpoints map[int]bool
}

// NewVM creates a new virtual machine
func NewVM() *VM {
	vm := &VM{
		Globals:         make(map[string]Value),
		NativeFunctions: make(map[string]*NativeFunction),
		OpenUpvalues:    make([]*Upvalue, 0),
		FrameIndex:      -1,
		Running:         false,
		Error:           nil,
		DebugMode:       false,
		Breakpoints:     make(map[int]bool),
	}
	
	// Initialize built-in functions
	vm.initBuiltins()
	
	return vm
}

// initBuiltins initializes built-in native functions
func (vm *VM) initBuiltins() {
	// Print function
	vm.RegisterNativeFunction("print", func(vm *VM, args []Value) (Value, error) {
		for i, arg := range args {
			if i > 0 {
				fmt.Print(" ")
			}
			fmt.Print(arg.ToString())
		}
		fmt.Println()
		return NilValue, nil
	}, 0, -1)
	
	// Type function
	vm.RegisterNativeFunction("type", func(vm *VM, args []Value) (Value, error) {
		if len(args) != 1 {
			return NilValue, NewRuntimeError("type() expects exactly 1 argument")
		}
		return NewStringValue(args[0].TypeName()), nil
	}, 1, 1)
	
	// Length function
	vm.RegisterNativeFunction("len", func(vm *VM, args []Value) (Value, error) {
		if len(args) != 1 {
			return NilValue, NewRuntimeError("len() expects exactly 1 argument")
		}
		
		arg := args[0]
		switch arg.Type {
		case TypeString:
			return NewIntValue(int64(len(arg.Data.(string)))), nil
		case TypeArray:
			return NewIntValue(int64(arg.Data.(*Array).Length())), nil
		case TypeObject:
			return NewIntValue(int64(len(arg.Data.(*Object).Properties))), nil
		default:
			return NilValue, NewRuntimeError("len() not supported for type %s", arg.TypeName())
		}
	}, 1, 1)
}

// RegisterNativeFunction registers a native function
func (vm *VM) RegisterNativeFunction(name string, fn NativeFunctionType, minArgs, maxArgs int) {
	vm.NativeFunctions[name] = NewNativeFunction(name, fn, minArgs, maxArgs)
}

// GetGlobal gets a global variable
func (vm *VM) GetGlobal(name string) (Value, bool) {
	val, ok := vm.Globals[name]
	return val, ok
}

// SetGlobal sets a global variable
func (vm *VM) SetGlobal(name string, value Value) {
	vm.Globals[name] = value
}

// GetRegister gets a register value
func (vm *VM) GetRegister(index int) Value {
	if index < 0 || index >= MaxRegisters {
		return NilValue
	}
	
	// Calculate actual register index based on current frame's base register
	actualIndex := index
	if vm.CurrentFrame != nil {
		actualIndex = vm.CurrentFrame.BaseReg + index
		if actualIndex >= MaxRegisters {
			return NilValue
		}
	}
	
	return vm.Registers[actualIndex]
}

// SetRegister sets a register value
func (vm *VM) SetRegister(index int, value Value) bool {
	if index < 0 || index >= MaxRegisters {
		return false
	}
	
	// Calculate actual register index based on current frame's base register
	actualIndex := index
	if vm.CurrentFrame != nil {
		actualIndex = vm.CurrentFrame.BaseReg + index
		if actualIndex >= MaxRegisters {
			return false
		}
	}
	
	vm.Registers[actualIndex] = value
	return true
}

// PushFrame pushes a new call frame
func (vm *VM) PushFrame(closure *Closure, baseReg, numRegs, returnAddr, numResults int) error {
	if vm.FrameIndex >= MaxFrames-1 {
		return NewVMErrorWithType(ErrStackOverflow, nil, "call stack overflow")
	}
	
	vm.FrameIndex++
	frame := &vm.Frames[vm.FrameIndex]
	frame.Closure = closure
	frame.PC = 0
	frame.BaseReg = baseReg
	frame.NumRegs = numRegs
	frame.ReturnAddr = returnAddr
	frame.NumResults = numResults
	
	vm.CurrentFrame = frame
	return nil
}

// PopFrame pops the current call frame
func (vm *VM) PopFrame() error {
	if vm.FrameIndex < 0 {
		return NewVMErrorWithType(ErrStackUnderflow, nil, "call stack underflow")
	}
	
	vm.FrameIndex--
	if vm.FrameIndex >= 0 {
		vm.CurrentFrame = &vm.Frames[vm.FrameIndex]
	} else {
		vm.CurrentFrame = nil
	}
	
	return nil
}

// Execute executes a function
func (vm *VM) Execute(closure *Closure, args []Value) (Value, error) {
	// Set up initial frame
	if err := vm.PushFrame(closure, 0, closure.Function.NumLocals, 0, 1); err != nil {
		return NilValue, err
	}
	
	// Copy arguments to registers
	for i, arg := range args {
		if i < closure.Function.NumParams {
			vm.SetRegister(i, arg)
		}
	}
	
	// Initialize remaining parameters to nil
	for i := len(args); i < closure.Function.NumParams; i++ {
		vm.SetRegister(i, NilValue)
	}
	
	vm.Running = true
	vm.Error = nil
	
	// Main execution loop
	for vm.Running && vm.Error == nil {
		if err := vm.executeInstruction(); err != nil {
			vm.Error = err
			break
		}
	}
	
	if vm.Error != nil {
		return NilValue, vm.Error
	}
	
	// Return the result
	if vm.CurrentFrame != nil && vm.CurrentFrame.ReturnAddr >= 0 {
		return vm.GetRegister(vm.CurrentFrame.ReturnAddr), nil
	}
	
	return NilValue, nil
}

// executeInstruction executes a single instruction
func (vm *VM) executeInstruction() error {
	if vm.CurrentFrame == nil {
		vm.Running = false
		return nil
	}
	
	frame := vm.CurrentFrame
	closure := frame.Closure
	
	// Check bounds
	if frame.PC >= len(closure.Function.Instructions) {
		vm.Running = false
		return nil
	}
	
	// Get instruction
	inst := closure.Function.Instructions[frame.PC]
	frame.PC++
	
	// Debug breakpoint
	if vm.DebugMode && vm.Breakpoints[frame.PC-1] {
		return NewRuntimeError("breakpoint at PC %d", frame.PC-1)
	}
	
	// Execute instruction
	return vm.executeOpCode(inst)
}

// executeOpCode executes a specific opcode
func (vm *VM) executeOpCode(inst Instruction) error {
	op := inst.GetOpCode()
	
	switch op {
	case OpMove:
		return vm.opMove(inst)
	case OpLoadK:
		return vm.opLoadK(inst)
	case OpLoadNil:
		return vm.opLoadNil(inst)
	case OpLoadBool:
		return vm.opLoadBool(inst)
	case OpLoadInt:
		return vm.opLoadInt(inst)
	case OpAdd:
		return vm.opAdd(inst)
	case OpSub:
		return vm.opSub(inst)
	case OpMul:
		return vm.opMul(inst)
	case OpDiv:
		return vm.opDiv(inst)
	case OpMod:
		return vm.opMod(inst)
	case OpNeg:
		return vm.opNeg(inst)
	case OpEq:
		return vm.opEq(inst)
	case OpNe:
		return vm.opNe(inst)
	case OpLt:
		return vm.opLt(inst)
	case OpLe:
		return vm.opLe(inst)
	case OpGt:
		return vm.opGt(inst)
	case OpGe:
		return vm.opGe(inst)
	case OpNot:
		return vm.opNot(inst)
	case OpAnd:
		return vm.opAnd(inst)
	case OpOr:
		return vm.opOr(inst)
	case OpJmp:
		return vm.opJmp(inst)
	case OpTest:
		return vm.opTest(inst)
	case OpCall:
		return vm.opCall(inst)
	case OpReturn:
		return vm.opReturn(inst)
	case OpNewTable:
		return vm.opNewTable(inst)
	case OpNewArray:
		return vm.opNewArray(inst)
	case OpGetTable:
		return vm.opGetTable(inst)
	case OpSetTable:
		return vm.opSetTable(inst)
	case OpGetGlobal:
		return vm.opGetGlobal(inst)
	case OpSetGlobal:
		return vm.opSetGlobal(inst)
	case OpHalt:
		vm.Running = false
		return nil
	case OpNop:
		return nil
	default:
		return NewRuntimeError("unknown opcode: %d", op)
	}
}

// Instruction implementations

func (vm *VM) opMove(inst Instruction) error {
	a, b := inst.GetA(), inst.GetB()
	vm.SetRegister(a, vm.GetRegister(b))
	return nil
}

func (vm *VM) opLoadK(inst Instruction) error {
	a, bx := inst.GetA(), inst.GetBx()
	if constant, ok := vm.CurrentFrame.Closure.Function.GetConstant(bx); ok {
		vm.SetRegister(a, constant)
	} else {
		return NewRuntimeError("invalid constant index: %d", bx)
	}
	return nil
}

func (vm *VM) opLoadNil(inst Instruction) error {
	a := inst.GetA()
	vm.SetRegister(a, NilValue)
	return nil
}

func (vm *VM) opLoadBool(inst Instruction) error {
	a, b, c := inst.GetA(), inst.GetB(), inst.GetC()
	vm.SetRegister(a, NewBoolValue(b != 0))
	if c != 0 {
		vm.CurrentFrame.PC++ // skip next instruction
	}
	return nil
}

func (vm *VM) opLoadInt(inst Instruction) error {
	a, bx := inst.GetA(), inst.GetBx()
	vm.SetRegister(a, NewIntValue(int64(bx-BxOffset)))
	return nil
}

func (vm *VM) opAdd(inst Instruction) error {
	a, b, c := inst.GetA(), inst.GetB(), inst.GetC()
	vb, vc := vm.GetRegister(b), vm.GetRegister(c)
	
	if vb.IsNumber() && vc.IsNumber() {
		if vb.IsInt() && vc.IsInt() {
			ib, _ := vb.ToInt()
			ic, _ := vc.ToInt()
			vm.SetRegister(a, NewIntValue(ib+ic))
		} else {
			fb, _ := vb.ToFloat()
			fc, _ := vc.ToFloat()
			vm.SetRegister(a, NewFloatValue(fb+fc))
		}
	} else if vb.IsString() && vc.IsString() {
		sb := vb.Data.(string)
		sc := vc.Data.(string)
		vm.SetRegister(a, NewStringValue(sb+sc))
	} else {
		return NewRuntimeError("cannot add %s and %s", vb.TypeName(), vc.TypeName())
	}
	
	return nil
}

func (vm *VM) opSub(inst Instruction) error {
	a, b, c := inst.GetA(), inst.GetB(), inst.GetC()
	vb, vc := vm.GetRegister(b), vm.GetRegister(c)
	
	if !vb.IsNumber() || !vc.IsNumber() {
		return NewRuntimeError("cannot subtract %s and %s", vb.TypeName(), vc.TypeName())
	}
	
	if vb.IsInt() && vc.IsInt() {
		ib, _ := vb.ToInt()
		ic, _ := vc.ToInt()
		vm.SetRegister(a, NewIntValue(ib-ic))
	} else {
		fb, _ := vb.ToFloat()
		fc, _ := vc.ToFloat()
		vm.SetRegister(a, NewFloatValue(fb-fc))
	}
	
	return nil
}

func (vm *VM) opMul(inst Instruction) error {
	a, b, c := inst.GetA(), inst.GetB(), inst.GetC()
	vb, vc := vm.GetRegister(b), vm.GetRegister(c)
	
	if !vb.IsNumber() || !vc.IsNumber() {
		return NewRuntimeError("cannot multiply %s and %s", vb.TypeName(), vc.TypeName())
	}
	
	if vb.IsInt() && vc.IsInt() {
		ib, _ := vb.ToInt()
		ic, _ := vc.ToInt()
		vm.SetRegister(a, NewIntValue(ib*ic))
	} else {
		fb, _ := vb.ToFloat()
		fc, _ := vc.ToFloat()
		vm.SetRegister(a, NewFloatValue(fb*fc))
	}
	
	return nil
}

func (vm *VM) opDiv(inst Instruction) error {
	a, b, c := inst.GetA(), inst.GetB(), inst.GetC()
	vb, vc := vm.GetRegister(b), vm.GetRegister(c)
	
	if !vb.IsNumber() || !vc.IsNumber() {
		return NewRuntimeError("cannot divide %s and %s", vb.TypeName(), vc.TypeName())
	}
	
	fb, _ := vb.ToFloat()
	fc, _ := vc.ToFloat()
	
	if fc == 0.0 {
		return NewVMErrorWithType(ErrDivisionByZero, nil, "division by zero")
	}
	
	vm.SetRegister(a, NewFloatValue(fb/fc))
	return nil
}

func (vm *VM) opMod(inst Instruction) error {
	a, b, c := inst.GetA(), inst.GetB(), inst.GetC()
	vb, vc := vm.GetRegister(b), vm.GetRegister(c)
	
	if !vb.IsNumber() || !vc.IsNumber() {
		return NewRuntimeError("cannot mod %s and %s", vb.TypeName(), vc.TypeName())
	}
	
	fb, _ := vb.ToFloat()
	fc, _ := vc.ToFloat()
	
	if fc == 0.0 {
		return NewVMErrorWithType(ErrDivisionByZero, nil, "modulo by zero")
	}
	
	vm.SetRegister(a, NewFloatValue(math.Mod(fb, fc)))
	return nil
}

func (vm *VM) opNeg(inst Instruction) error {
	a, b := inst.GetA(), inst.GetB()
	vb := vm.GetRegister(b)
	
	if !vb.IsNumber() {
		return NewRuntimeError("cannot negate %s", vb.TypeName())
	}
	
	if vb.IsInt() {
		ib, _ := vb.ToInt()
		vm.SetRegister(a, NewIntValue(-ib))
	} else {
		fb, _ := vb.ToFloat()
		vm.SetRegister(a, NewFloatValue(-fb))
	}
	
	return nil
}

func (vm *VM) opEq(inst Instruction) error {
	a, b, c := inst.GetA(), inst.GetB(), inst.GetC()
	vb, vc := vm.GetRegister(b), vm.GetRegister(c)
	
	result := vb.Equals(vc)
	vm.SetRegister(a, NewBoolValue(result))
	
	return nil
}

func (vm *VM) opNe(inst Instruction) error {
	a, b, c := inst.GetA(), inst.GetB(), inst.GetC()
	vb, vc := vm.GetRegister(b), vm.GetRegister(c)
	
	result := !vb.Equals(vc)
	vm.SetRegister(a, NewBoolValue(result))
	
	return nil
}

func (vm *VM) opLt(inst Instruction) error {
	a, b, c := inst.GetA(), inst.GetB(), inst.GetC()
	vb, vc := vm.GetRegister(b), vm.GetRegister(c)
	
	result := false
	if cmp, ok := vb.Compare(vc); ok {
		result = cmp < 0
	}
	vm.SetRegister(a, NewBoolValue(result))
	
	return nil
}

func (vm *VM) opLe(inst Instruction) error {
	a, b, c := inst.GetA(), inst.GetB(), inst.GetC()
	vb, vc := vm.GetRegister(b), vm.GetRegister(c)
	
	result := false
	if cmp, ok := vb.Compare(vc); ok {
		result = cmp <= 0
	}
	vm.SetRegister(a, NewBoolValue(result))
	
	return nil
}

func (vm *VM) opGt(inst Instruction) error {
	a, b, c := inst.GetA(), inst.GetB(), inst.GetC()
	vb, vc := vm.GetRegister(b), vm.GetRegister(c)
	
	result := false
	if cmp, ok := vb.Compare(vc); ok {
		result = cmp > 0
	}
	vm.SetRegister(a, NewBoolValue(result))
	
	return nil
}

func (vm *VM) opGe(inst Instruction) error {
	a, b, c := inst.GetA(), inst.GetB(), inst.GetC()
	vb, vc := vm.GetRegister(b), vm.GetRegister(c)
	
	result := false
	if cmp, ok := vb.Compare(vc); ok {
		result = cmp >= 0
	}
	vm.SetRegister(a, NewBoolValue(result))
	
	return nil
}

func (vm *VM) opNot(inst Instruction) error {
	a, b := inst.GetA(), inst.GetB()
	vb := vm.GetRegister(b)
	vm.SetRegister(a, NewBoolValue(!vb.ToBool()))
	return nil
}

func (vm *VM) opJmp(inst Instruction) error {
	bx := inst.GetBx()
	vm.CurrentFrame.PC += bx - BxOffset
	return nil
}

func (vm *VM) opTest(inst Instruction) error {
	a := inst.GetA()
	va := vm.GetRegister(a)
	
	if !va.ToBool() {
		vm.CurrentFrame.PC++ // skip next instruction
	}
	
	return nil
}

func (vm *VM) opCall(inst Instruction) error {
	a, b, c := inst.GetA(), inst.GetB(), inst.GetC()
	
	// Get function to call
	fn := vm.GetRegister(a)
	
	// Collect arguments
	args := make([]Value, b)
	for i := 0; i < b; i++ {
		args[i] = vm.GetRegister(a + 1 + i)
	}
	
	// Call function
	if fn.Type == TypeNativeFunction {
		nativeFn := fn.Data.(*NativeFunction)
		result, err := nativeFn.Call(vm, args)
		if err != nil {
			return err
		}
		
		// Store result
		if c > 0 {
			vm.SetRegister(a, result)
		}
	} else if fn.Type == TypeFunction {
		// User-defined function call
		function := fn.Data.(*Function)
		
		// Create closure for the function
		closure := NewClosure(function)
		
		// Check argument count
		if len(args) < function.NumParams {
			return NewRuntimeError("function '%s' expects %d arguments, got %d", 
				function.Name, function.NumParams, len(args))
		}
		
		// Push new call frame
		returnAddr := -1
		if c > 0 {
			returnAddr = a
		}
		
		// Calculate base register for new frame
		// Each frame needs its own register space
		newBaseReg := 0
		if vm.CurrentFrame != nil {
			newBaseReg = vm.CurrentFrame.BaseReg + vm.CurrentFrame.NumRegs
		}
		
		if err := vm.PushFrame(closure, newBaseReg, function.NumLocals, returnAddr, c); err != nil {
			return err
		}
		
		// Copy arguments to registers
		for i, arg := range args {
			if i < function.NumParams {
				vm.SetRegister(i, arg)
			}
		}
		
		// Initialize remaining parameters to nil
		for i := len(args); i < function.NumParams; i++ {
			vm.SetRegister(i, NilValue)
		}
	} else {
		return NewRuntimeError("attempt to call %s value", fn.TypeName())
	}
	
	return nil
}

func (vm *VM) opReturn(inst Instruction) error {
	a, b := inst.GetA(), inst.GetB()
	
	// Get return value before popping frame
	var returnValue Value = NilValue
	if b > 0 {
		returnValue = vm.GetRegister(a)
	}
	
	// Store return address before popping frame
	returnAddr := vm.CurrentFrame.ReturnAddr
	
	// Pop frame
	if err := vm.PopFrame(); err != nil {
		return err
	}
	
	// Copy return value to caller's frame
	if b > 0 && returnAddr >= 0 {
		vm.SetRegister(returnAddr, returnValue)
	}
	
	// If no more frames, stop execution
	if vm.CurrentFrame == nil {
		vm.Running = false
	}
	
	return nil
}

func (vm *VM) opNewTable(inst Instruction) error {
	a := inst.GetA()
	obj := NewObject()
	vm.SetRegister(a, NewObjectValue(obj))
	return nil
}

func (vm *VM) opNewArray(inst Instruction) error {
	a, bx := inst.GetA(), inst.GetBx()
	arr := NewArray(bx)
	vm.SetRegister(a, NewArrayValue(arr))
	return nil
}

func (vm *VM) opGetTable(inst Instruction) error {
	a, b, c := inst.GetA(), inst.GetB(), inst.GetC()
	table := vm.GetRegister(b)
	key := vm.GetRegister(c)
	
	if table.Type == TypeObject && key.Type == TypeString {
		obj := table.Data.(*Object)
		keyStr := key.Data.(string)
		if val, ok := obj.Get(keyStr); ok {
			vm.SetRegister(a, val)
		} else {
			vm.SetRegister(a, NilValue)
		}
	} else if table.Type == TypeArray && key.Type == TypeInt {
		arr := table.Data.(*Array)
		index, _ := key.ToInt()
		if val, ok := arr.Get(int(index)); ok {
			vm.SetRegister(a, val)
		} else {
			vm.SetRegister(a, NilValue)
		}
	} else {
		return NewRuntimeError("invalid table access: %s[%s]", table.TypeName(), key.TypeName())
	}
	
	return nil
}

func (vm *VM) opSetTable(inst Instruction) error {
	a, b, c := inst.GetA(), inst.GetB(), inst.GetC()
	table := vm.GetRegister(a)
	key := vm.GetRegister(b)
	value := vm.GetRegister(c)
	
	if table.Type == TypeObject && key.Type == TypeString {
		obj := table.Data.(*Object)
		keyStr := key.Data.(string)
		obj.Set(keyStr, value)
	} else if table.Type == TypeArray && key.Type == TypeInt {
		arr := table.Data.(*Array)
		index, _ := key.ToInt()
		arr.Set(int(index), value)
	} else {
		return NewRuntimeError("invalid table assignment: %s[%s]", table.TypeName(), key.TypeName())
	}
	
	return nil
}

func (vm *VM) opGetGlobal(inst Instruction) error {
	a, bx := inst.GetA(), inst.GetBx()
	
	// Get constant (should be string)
	constant, ok := vm.CurrentFrame.Closure.Function.GetConstant(bx)
	if !ok || constant.Type != TypeString {
		return NewRuntimeError("invalid global name constant")
	}
	
	name := constant.Data.(string)
	
	// Check native functions first
	if nativeFn, ok := vm.NativeFunctions[name]; ok {
		vm.SetRegister(a, NewNativeFunctionValue(nativeFn))
		return nil
	}
	
	// Check global variables
	if val, ok := vm.GetGlobal(name); ok {
		vm.SetRegister(a, val)
	} else {
		return NewVMErrorWithType(ErrUndefinedVariable, nil, "undefined variable: %s", name)
	}
	
	return nil
}

func (vm *VM) opSetGlobal(inst Instruction) error {
	a, bx := inst.GetA(), inst.GetBx()
	
	// Get constant (should be string)
	constant, ok := vm.CurrentFrame.Closure.Function.GetConstant(bx)
	if !ok || constant.Type != TypeString {
		return NewRuntimeError("invalid global name constant")
	}
	
	name := constant.Data.(string)
	value := vm.GetRegister(a)
	
	vm.SetGlobal(name, value)
	return nil
}

func (vm *VM) opAnd(inst Instruction) error {
	a, b, c := inst.GetA(), inst.GetB(), inst.GetC()
	vb, vc := vm.GetRegister(b), vm.GetRegister(c)
	
	// Logical AND: result is true only if both operands are true
	result := vb.ToBool() && vc.ToBool()
	vm.SetRegister(a, NewBoolValue(result))
	
	return nil
}

func (vm *VM) opOr(inst Instruction) error {
	a, b, c := inst.GetA(), inst.GetB(), inst.GetC()
	vb, vc := vm.GetRegister(b), vm.GetRegister(c)
	
	// Logical OR: result is true if either operand is true
	result := vb.ToBool() || vc.ToBool()
	vm.SetRegister(a, NewBoolValue(result))
	
	return nil
}