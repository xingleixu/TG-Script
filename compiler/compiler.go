package compiler

import (
	"fmt"

	"github.com/xingleixu/TG-Script/ast"
	"github.com/xingleixu/TG-Script/vm"
)

// Compiler compiles AST to VM instructions
type Compiler struct {
	function     *vm.Function
	symbolTable  *SymbolTable
	nextRegister int
	maxRegisters int
	constants    []vm.Value
	instructions []vm.Instruction
	errors       []error
}

// SymbolTable manages variable scoping
type SymbolTable struct {
	parent  *SymbolTable
	symbols map[string]*Symbol
	level   int
}

// Symbol represents a variable or function
type Symbol struct {
	Name     string
	Type     SymbolType
	Register int
	Level    int
}

// SymbolType represents the type of symbol
type SymbolType int

const (
	SymbolLocal SymbolType = iota
	SymbolGlobal
	SymbolFunction
	SymbolBuiltin
)

// NewCompiler creates a new compiler
func NewCompiler() *Compiler {
	return &Compiler{
		function:     vm.NewFunction("main"),
		symbolTable:  NewSymbolTable(nil),
		nextRegister: 0,
		maxRegisters: 0,
		constants:    make([]vm.Value, 0),
		instructions: make([]vm.Instruction, 0),
		errors:       make([]error, 0),
	}
}

// CompileFunction compiles a program to a function
func CompileFunction(program *ast.Program) (*vm.Function, error) {
	compiler := NewCompiler()
	
	if err := compiler.compileProgram(program); err != nil {
		return nil, err
	}
	
	if compiler.HasErrors() {
		return nil, fmt.Errorf("compilation errors: %v", compiler.GetErrors())
	}
	
	return compiler.GetFunction(), nil
}

// NewSymbolTable creates a new symbol table
func NewSymbolTable(parent *SymbolTable) *SymbolTable {
	level := 0
	if parent != nil {
		level = parent.level + 1
	}
	
	return &SymbolTable{
		parent:  parent,
		symbols: make(map[string]*Symbol),
		level:   level,
	}
}

// Define defines a symbol in the current scope
func (st *SymbolTable) Define(name string, symbolType SymbolType, register int) *Symbol {
	symbol := &Symbol{
		Name:     name,
		Type:     symbolType,
		Register: register,
		Level:    st.level,
	}
	st.symbols[name] = symbol
	return symbol
}

// Resolve resolves a symbol by name
func (st *SymbolTable) Resolve(name string) (*Symbol, bool) {
	symbol, ok := st.symbols[name]
	if ok {
		return symbol, true
	}
	
	if st.parent != nil {
		return st.parent.Resolve(name)
	}
	
	return nil, false
}

// AllocateRegister allocates a new register
func (c *Compiler) AllocateRegister() int {
	reg := c.nextRegister
	c.nextRegister++
	if c.nextRegister > c.maxRegisters {
		c.maxRegisters = c.nextRegister
	}
	return reg
}

// FreeRegister frees a register (simplified implementation)
func (c *Compiler) FreeRegister(reg int) {
	// In a more sophisticated implementation, we would track free registers
}

// AddConstant adds a constant to the constants pool
func (c *Compiler) AddConstant(value vm.Value) int {
	// Check if constant already exists
	for i, constant := range c.constants {
		if constant.Type == value.Type {
			switch value.Type {
			case vm.TypeInt:
				if constant.Data.(int64) == value.Data.(int64) {
					return i
				}
			case vm.TypeString:
				if constant.Data.(string) == value.Data.(string) {
					return i
				}
			}
		}
	}
	
	// Add new constant
	c.constants = append(c.constants, value)
	return len(c.constants) - 1
}

// Emit emits an instruction
func (c *Compiler) Emit(opcode vm.OpCode, operands ...int) int {
	var inst vm.Instruction
	
	switch len(operands) {
	case 0:
		inst = vm.CreateABC(opcode, 0, 0, 0)
	case 1:
		inst = vm.CreateABC(opcode, operands[0], 0, 0)
	case 2:
		inst = vm.CreateABx(opcode, operands[0], operands[1])
	case 3:
		inst = vm.CreateABC(opcode, operands[0], operands[1], operands[2])
	default:
		c.AddError(fmt.Errorf("too many operands for instruction: %d", len(operands)))
		return len(c.instructions)
	}
	
	c.instructions = append(c.instructions, inst)
	return len(c.instructions) - 1
}

// PatchJump patches a jump instruction
func (c *Compiler) PatchJump(pos int, target int) {
	if pos >= len(c.instructions) {
		c.AddError(fmt.Errorf("invalid jump position: %d", pos))
		return
	}
	
	inst := c.instructions[pos]
	opcode := inst.GetOpCode()
	a := inst.GetA()
	
	// Calculate relative offset from the instruction after the jump
	offset := target - (pos + 1)
	
	// Create new instruction with the offset + BxOffset for signed Bx
	c.instructions[pos] = vm.CreateABx(opcode, a, offset + vm.BxOffset)
}

// AddError adds an error to the error list
func (c *Compiler) AddError(err error) {
	c.errors = append(c.errors, err)
}

// HasErrors returns true if there are compilation errors
func (c *Compiler) HasErrors() bool {
	return len(c.errors) > 0
}

// GetErrors returns the list of compilation errors
func (c *Compiler) GetErrors() []error {
	return c.errors
}

// compileProgram compiles a program
func (c *Compiler) compileProgram(program *ast.Program) error {
	for _, stmt := range program.Body {
		if err := c.compileStatement(stmt); err != nil {
			return err
		}
	}
	
	// Emit halt instruction
	c.Emit(vm.OpHalt)
	
	// Finalize function
	c.function.Instructions = c.instructions
	c.function.Constants = c.constants
	c.function.NumLocals = c.maxRegisters
	
	return nil
}

// compileStatement compiles a statement
func (c *Compiler) compileStatement(stmt ast.Statement) error {
	switch s := stmt.(type) {
	case *ast.ExpressionStatement:
		return c.compileExpressionStatement(s)
	case *ast.VariableDeclaration:
		return c.compileVariableDeclaration(s)
	case *ast.IfStatement:
		return c.compileIfStatement(s)
	case *ast.ForStatement:
		return c.compileForStatement(s)
	case *ast.WhileStatement:
		return c.compileWhileStatement(s)
	case *ast.ReturnStatement:
		return c.compileReturnStatement(s)
	case *ast.BlockStatement:
		return c.compileBlockStatement(s)
	default:
		return fmt.Errorf("unsupported statement type: %T", stmt)
	}
}

// compileExpressionStatement compiles an expression statement
func (c *Compiler) compileExpressionStatement(stmt *ast.ExpressionStatement) error {
	reg := c.AllocateRegister()
	defer c.FreeRegister(reg)
	
	return c.compileExpression(stmt.Expression, reg)
}

// compileVariableDeclaration compiles a variable declaration
func (c *Compiler) compileVariableDeclaration(stmt *ast.VariableDeclaration) error {
	for _, decl := range stmt.Declarations {
		reg := c.AllocateRegister()
		
		// Compile initializer if present
		if decl.Init != nil {
			if err := c.compileExpression(decl.Init, reg); err != nil {
				return err
			}
		} else {
			// Initialize to nil
			c.Emit(vm.OpLoadNil, reg)
		}
		
		// Define symbol - handle BindingTarget properly
		if id, ok := decl.Id.(*ast.Identifier); ok {
			c.symbolTable.Define(id.Name, SymbolLocal, reg)
		}
	}
	
	return nil
}

// compileIfStatement compiles an if statement
func (c *Compiler) compileIfStatement(stmt *ast.IfStatement) error {
	// Compile condition
	condReg := c.AllocateRegister()
	if err := c.compileExpression(stmt.Test, condReg); err != nil {
		return err
	}
	
	// Test condition and jump if false
	c.Emit(vm.OpTest, condReg)
	jumpToElse := c.Emit(vm.OpJmp, 0) // placeholder
	c.FreeRegister(condReg)
	
	// Compile then branch
	if err := c.compileStatement(stmt.Consequent); err != nil {
		return err
	}
	
	if stmt.Alternate != nil {
		// Jump over else branch
		jumpToEnd := c.Emit(vm.OpJmp, 0) // placeholder
		
		// Patch jump to else
		c.PatchJump(jumpToElse, len(c.instructions))
		
		// Compile else branch
		if err := c.compileStatement(stmt.Alternate); err != nil {
			return err
		}
		
		// Patch jump to end
		c.PatchJump(jumpToEnd, len(c.instructions))
	} else {
		// Patch jump to end
		c.PatchJump(jumpToElse, len(c.instructions))
	}
	
	return nil
}

// compileReturnStatement compiles a return statement
func (c *Compiler) compileReturnStatement(stmt *ast.ReturnStatement) error {
	if stmt.Argument != nil {
		// Compile return value
		reg := c.AllocateRegister()
		if err := c.compileExpression(stmt.Argument, reg); err != nil {
			return err
		}
		
		// Return with value
		c.Emit(vm.OpReturn, reg)
		c.FreeRegister(reg)
	} else {
		// Return nil
		c.Emit(vm.OpReturn, 0)
	}
	
	return nil
}

// compileBlockStatement compiles a block statement
func (c *Compiler) compileBlockStatement(block *ast.BlockStatement) error {
	// Enter new scope
	c.symbolTable = NewSymbolTable(c.symbolTable)
	
	for _, stmt := range block.Body {
		if err := c.compileStatement(stmt); err != nil {
			return err
		}
	}
	
	// Exit scope
	c.symbolTable = c.symbolTable.parent
	
	return nil
}

// compileExpression compiles an expression
func (c *Compiler) compileExpression(expr ast.Expression, targetReg int) error {
	switch e := expr.(type) {
	case *ast.Identifier:
		return c.compileIdentifier(e, targetReg)
	case *ast.IntegerLiteral:
		return c.compileIntegerLiteral(e, targetReg)
	case *ast.StringLiteral:
		return c.compileStringLiteral(e, targetReg)
	case *ast.BooleanLiteral:
		return c.compileBooleanLiteral(e, targetReg)
	case *ast.NullLiteral:
		return c.compileNullLiteral(e, targetReg)
	case *ast.UndefinedLiteral:
		return c.compileUndefinedLiteral(e, targetReg)
	case *ast.VoidLiteral:
		return c.compileVoidLiteral(e, targetReg)
	case *ast.ArrayLiteral:
		return c.compileArrayLiteral(e, targetReg)
	case *ast.BinaryExpression:
		return c.compileBinaryExpression(e, targetReg)
	case *ast.UnaryExpression:
		return c.compileUnaryExpression(e, targetReg)
	case *ast.CallExpression:
		return c.compileCallExpression(e, targetReg)
	case *ast.AssignmentExpression:
		return c.compileAssignmentExpression(e, targetReg)
	case *ast.MemberExpression:
		return c.compileMemberExpression(e, targetReg)
	default:
		return fmt.Errorf("unsupported expression type: %T", expr)
	}
}

// compileIdentifier compiles an identifier
func (c *Compiler) compileIdentifier(expr *ast.Identifier, targetReg int) error {
	symbol, ok := c.symbolTable.Resolve(expr.Name)
	if !ok {
		// Try to load as global
		constIndex := c.AddConstant(vm.NewStringValue(expr.Name))
		c.Emit(vm.OpGetGlobal, targetReg, constIndex)
		return nil
	}
	
	if symbol.Type == SymbolLocal {
		c.Emit(vm.OpMove, targetReg, symbol.Register)
	} else {
		// Global variable
		constIndex := c.AddConstant(vm.NewStringValue(expr.Name))
		c.Emit(vm.OpGetGlobal, targetReg, constIndex)
	}
	
	return nil
}

// compileIntegerLiteral compiles an integer literal
func (c *Compiler) compileIntegerLiteral(expr *ast.IntegerLiteral, targetReg int) error {
	constIndex := c.AddConstant(vm.NewIntValue(expr.Value))
	c.Emit(vm.OpLoadK, targetReg, constIndex)
	return nil
}

// compileStringLiteral compiles a string literal
func (c *Compiler) compileStringLiteral(expr *ast.StringLiteral, targetReg int) error {
	constIndex := c.AddConstant(vm.NewStringValue(expr.Value))
	c.Emit(vm.OpLoadK, targetReg, constIndex)
	return nil
}

// compileBooleanLiteral compiles a boolean literal
func (c *Compiler) compileBooleanLiteral(expr *ast.BooleanLiteral, targetReg int) error {
	if expr.Value {
		c.Emit(vm.OpLoadBool, targetReg, 1, 0)
	} else {
		c.Emit(vm.OpLoadBool, targetReg, 0, 0)
	}
	return nil
}

// compileNullLiteral compiles a null literal
func (c *Compiler) compileNullLiteral(expr *ast.NullLiteral, targetReg int) error {
	constIndex := c.AddConstant(vm.NullValue)
	c.Emit(vm.OpLoadK, targetReg, constIndex)
	return nil
}

// compileUndefinedLiteral compiles an undefined literal
func (c *Compiler) compileUndefinedLiteral(expr *ast.UndefinedLiteral, targetReg int) error {
	constIndex := c.AddConstant(vm.NilValue)
	c.Emit(vm.OpLoadK, targetReg, constIndex)
	return nil
}

// compileVoidLiteral compiles a void literal
func (c *Compiler) compileVoidLiteral(expr *ast.VoidLiteral, targetReg int) error {
	constIndex := c.AddConstant(vm.VoidValue)
	c.Emit(vm.OpLoadK, targetReg, constIndex)
	return nil
}

// compileArrayLiteral compiles an array literal
func (c *Compiler) compileArrayLiteral(expr *ast.ArrayLiteral, targetReg int) error {
	// Create new array with capacity equal to number of elements
	c.Emit(vm.OpNewArray, targetReg, len(expr.Elements))
	
	// Compile and set each element
	for i, element := range expr.Elements {
		if element != nil {
			// Compile element value
			valueReg := c.AllocateRegister()
			if err := c.compileExpression(element, valueReg); err != nil {
				return err
			}
			
			// Create index value
			indexReg := c.AllocateRegister()
			constIndex := c.AddConstant(vm.NewIntValue(int64(i)))
			c.Emit(vm.OpLoadK, indexReg, constIndex)
			
			// Set array element: array[index] = value
			c.Emit(vm.OpSetTable, targetReg, indexReg, valueReg)
			
			c.FreeRegister(valueReg)
			c.FreeRegister(indexReg)
		}
		// For nil elements (holes in sparse arrays), we don't need to do anything
		// as the array will have nil values by default
	}
	
	return nil
}

// compileBinaryExpression compiles a binary expression
func (c *Compiler) compileUnaryExpression(expr *ast.UnaryExpression, targetReg int) error {
	operandReg := c.AllocateRegister()
	defer c.FreeRegister(operandReg)

	if err := c.compileExpression(expr.Operand, operandReg); err != nil {
		return err
	}

	switch expr.Operator.String() {
	case "!":
		c.Emit(vm.OpNot, targetReg, operandReg)
	case "-":
		// For negative numbers, we can use subtraction from 0
		zeroReg := c.AllocateRegister()
		defer c.FreeRegister(zeroReg)
		c.Emit(vm.OpLoadK, zeroReg, c.AddConstant(vm.NewIntValue(0)))
		c.Emit(vm.OpSub, targetReg, zeroReg, operandReg)
	default:
		return fmt.Errorf("unsupported unary operator: %s", expr.Operator.String())
	}

	return nil
}

func (c *Compiler) compileBinaryExpression(expr *ast.BinaryExpression, targetReg int) error {
	// Compile operands
	leftReg := c.AllocateRegister()
	rightReg := c.AllocateRegister()
	
	if err := c.compileExpression(expr.Left, leftReg); err != nil {
		return err
	}
	
	if err := c.compileExpression(expr.Right, rightReg); err != nil {
		return err
	}
	
	// Emit operation based on operator
	switch expr.Operator.String() {
	case "+":
		c.Emit(vm.OpAdd, targetReg, leftReg, rightReg)
	case "-":
		c.Emit(vm.OpSub, targetReg, leftReg, rightReg)
	case "*":
		c.Emit(vm.OpMul, targetReg, leftReg, rightReg)
	case "/":
		c.Emit(vm.OpDiv, targetReg, leftReg, rightReg)
	case "%":
		c.Emit(vm.OpMod, targetReg, leftReg, rightReg)
	case "==":
		c.Emit(vm.OpEq, targetReg, leftReg, rightReg)
	case "!=":
		c.Emit(vm.OpNe, targetReg, leftReg, rightReg)
	case "<":
		c.Emit(vm.OpLt, targetReg, leftReg, rightReg)
	case "<=":
		c.Emit(vm.OpLe, targetReg, leftReg, rightReg)
	case ">":
		c.Emit(vm.OpGt, targetReg, leftReg, rightReg)
	case ">=":
		c.Emit(vm.OpGe, targetReg, leftReg, rightReg)
	case "&&":
		c.Emit(vm.OpAnd, targetReg, leftReg, rightReg)
	case "||":
		c.Emit(vm.OpOr, targetReg, leftReg, rightReg)
	default:
		return fmt.Errorf("unsupported binary operator: %s", expr.Operator.String())
	}
	
	c.FreeRegister(leftReg)
	c.FreeRegister(rightReg)
	
	return nil
}

// compileCallExpression compiles a function call expression
func (c *Compiler) compileCallExpression(expr *ast.CallExpression, targetReg int) error {
	// Compile the function being called
	funcReg := c.AllocateRegister()
	if err := c.compileExpression(expr.Callee, funcReg); err != nil {
		return err
	}
	
	// Compile arguments
	argRegs := make([]int, len(expr.Arguments))
	for i, arg := range expr.Arguments {
		argReg := c.AllocateRegister()
		if err := c.compileExpression(arg, argReg); err != nil {
			return err
		}
		argRegs[i] = argReg
	}
	
	// Move function to target register
	c.Emit(vm.OpMove, targetReg, funcReg)
	
	// Move arguments to consecutive registers after function
	for i, argReg := range argRegs {
		c.Emit(vm.OpMove, targetReg+1+i, argReg)
	}
	
	// Emit call instruction
	// OpCall format: R(A)..R(A+C-1) := R(A)(R(A+1)..R(A+B-1))
	// A = target register (where result goes)
	// B = number of arguments + 1
	// C = number of return values + 1
	c.Emit(vm.OpCall, targetReg, len(expr.Arguments)+1, 1)
	
	return nil
}

// compileAssignmentExpression compiles an assignment expression
func (c *Compiler) compileAssignmentExpression(expr *ast.AssignmentExpression, targetReg int) error {
	// For now, only support simple assignment (=)
	if expr.Operator.String() != "=" {
		return fmt.Errorf("unsupported assignment operator: %s", expr.Operator.String())
	}
	
	// Compile the right-hand side first
	valueReg := c.AllocateRegister()
	defer c.FreeRegister(valueReg)
	
	if err := c.compileExpression(expr.Right, valueReg); err != nil {
		return err
	}
	
	// Handle left-hand side assignment
	switch left := expr.Left.(type) {
	case *ast.Identifier:
		// Simple variable assignment
		symbol, exists := c.symbolTable.Resolve(left.Name)
		if exists && symbol.Type == SymbolLocal {
			// Local variable assignment
			c.Emit(vm.OpMove, symbol.Register, valueReg)
			c.Emit(vm.OpMove, targetReg, symbol.Register)
		} else {
			// Global variable assignment
			constIndex := c.AddConstant(vm.NewStringValue(left.Name))
			c.Emit(vm.OpSetGlobal, valueReg, constIndex)
			c.Emit(vm.OpMove, targetReg, valueReg)
		}
		return nil
		
	case *ast.MemberExpression:
		// Member expression assignment (obj[prop] = value)
		objReg := c.AllocateRegister()
		defer c.FreeRegister(objReg)
		
		if err := c.compileExpression(left.Object, objReg); err != nil {
			return err
		}
		
		propReg := c.AllocateRegister()
		defer c.FreeRegister(propReg)
		
		if err := c.compileExpression(left.Property, propReg); err != nil {
			return err
		}
		
		// Emit OpSetTable instruction to set the value
		c.Emit(vm.OpSetTable, objReg, propReg, valueReg)
		c.Emit(vm.OpMove, targetReg, valueReg)
		return nil
		
	default:
		return fmt.Errorf("unsupported assignment target: %T", expr.Left)
	}
}

// compileForStatement compiles a for statement
func (c *Compiler) compileForStatement(stmt *ast.ForStatement) error {
	// Enter new scope for loop variables
	c.symbolTable = NewSymbolTable(c.symbolTable)
	defer func() {
		c.symbolTable = c.symbolTable.parent
	}()
	
	// Compile initialization if present
	if stmt.Init != nil {
		if err := c.compileStatement(stmt.Init); err != nil {
			return err
		}
	}
	
	// Loop start position
	loopStart := len(c.instructions)
	
	// Compile test condition if present
	var jumpToEnd int
	if stmt.Test != nil {
		condReg := c.AllocateRegister()
		if err := c.compileExpression(stmt.Test, condReg); err != nil {
			return err
		}
		
		// Test condition and jump if false
		c.Emit(vm.OpTest, condReg)
		jumpToEnd = c.Emit(vm.OpJmp, 0) // placeholder
		c.FreeRegister(condReg)
	}
	
	// Compile body
	if err := c.compileStatement(stmt.Body); err != nil {
		return err
	}
	
	// Compile update if present
	if stmt.Update != nil {
		reg := c.AllocateRegister()
		defer c.FreeRegister(reg)
		if err := c.compileExpression(stmt.Update, reg); err != nil {
			return err
		}
	}
	
	// Jump back to loop start
	offset := loopStart - (len(c.instructions) + 1)
	c.Emit(vm.OpJmp, offset + vm.BxOffset)
	
	// Patch jump to end if test condition exists
	if stmt.Test != nil {
		c.PatchJump(jumpToEnd, len(c.instructions))
	}
	
	return nil
}

// compileWhileStatement compiles a while statement
func (c *Compiler) compileWhileStatement(stmt *ast.WhileStatement) error {
	// Loop start position
	loopStart := len(c.instructions)
	
	// Compile test condition
	condReg := c.AllocateRegister()
	if err := c.compileExpression(stmt.Test, condReg); err != nil {
		return err
	}
	
	// Test condition and jump if false
	c.Emit(vm.OpTest, condReg)
	jumpToEnd := c.Emit(vm.OpJmp, 0) // placeholder
	c.FreeRegister(condReg)
	
	// Compile body
	if err := c.compileStatement(stmt.Body); err != nil {
		return err
	}
	
	// Jump back to loop start
	offset := loopStart - (len(c.instructions) + 1)
	c.Emit(vm.OpJmp, offset + vm.BxOffset)
	
	// Patch jump to end
	c.PatchJump(jumpToEnd, len(c.instructions))
	
	return nil
}

// compileMemberExpression compiles a member expression (obj[prop] or obj.prop)
func (c *Compiler) compileMemberExpression(expr *ast.MemberExpression, targetReg int) error {
	// Compile the object
	objReg := c.AllocateRegister()
	if err := c.compileExpression(expr.Object, objReg); err != nil {
		return err
	}
	defer c.FreeRegister(objReg)

	// Compile the property/index
	propReg := c.AllocateRegister()
	if err := c.compileExpression(expr.Property, propReg); err != nil {
		return err
	}
	defer c.FreeRegister(propReg)

	// Emit OpGetTable instruction to get the value
	c.Emit(vm.OpGetTable, targetReg, objReg, propReg)
	
	return nil
}

// GetFunction returns the compiled function
func (c *Compiler) GetFunction() *vm.Function {
	return c.function
}