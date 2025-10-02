package vm

import "fmt"

// OpCode represents a virtual machine instruction opcode
type OpCode byte

// Instruction opcodes
const (
	// Data movement
	OpMove     OpCode = iota // R(A) := R(B)
	OpLoadK                  // R(A) := K(Bx)
	OpLoadNil                // R(A) := nil
	OpLoadBool               // R(A) := bool(B); if C then PC++
	OpLoadInt                // R(A) := int(sBx)

	// Arithmetic operations
	OpAdd // R(A) := R(B) + R(C)
	OpSub // R(A) := R(B) - R(C)
	OpMul // R(A) := R(B) * R(C)
	OpDiv // R(A) := R(B) / R(C)
	OpMod // R(A) := R(B) % R(C)
	OpPow // R(A) := R(B) ^ R(C)
	OpNeg // R(A) := -R(B)

	// Bitwise operations
	OpBitAnd // R(A) := R(B) & R(C)
	OpBitOr  // R(A) := R(B) | R(C)
	OpBitXor // R(A) := R(B) ^ R(C)
	OpBitNot // R(A) := ~R(B)
	OpShl    // R(A) := R(B) << R(C)
	OpShr    // R(A) := R(B) >> R(C)

	// Comparison operations
	OpEq // if R(B) == R(C) then PC++
	OpNe // if R(B) != R(C) then PC++
	OpLt // if R(B) < R(C) then PC++
	OpLe // if R(B) <= R(C) then PC++
	OpGt // if R(B) > R(C) then PC++
	OpGe // if R(B) >= R(C) then PC++

	// Logical operations
	OpNot // R(A) := !R(B)
	OpAnd // R(A) := R(B) && R(C)
	OpOr  // R(A) := R(B) || R(C)

	// Control flow
	OpJmp     // PC += sBx
	OpTest    // if not R(A) then PC++
	OpTestSet // if R(B) then R(A) := R(B) else PC++

	// Function calls
	OpCall     // R(A)..R(A+C-1) := R(A)(R(A+1)..R(A+B-1))
	OpTailCall // return R(A)(R(A+1)..R(A+B-1))
	OpReturn   // return R(A)..R(A+B-1)

	// Object operations
	OpNewTable  // R(A) := {} (size = B*C)
	OpGetTable  // R(A) := R(B)[R(C)]
	OpSetTable  // R(A)[R(B)] := R(C)
	OpGetGlobal // R(A) := G[K(Bx)]
	OpSetGlobal // G[K(Bx)] := R(A)
	OpGetUpval  // R(A) := UpValue[B]
	OpSetUpval  // UpValue[B] := R(A)

	// Array operations
	OpNewArray // R(A) := [] (size = Bx)
	OpGetIndex // R(A) := R(B)[R(C)]
	OpSetIndex // R(A)[R(B)] := R(C)
	OpLen      // R(A) := len(R(B))

	// String operations
	OpConcat // R(A) := R(B) .. R(C)

	// Type operations
	OpTypeOf   // R(A) := typeof(R(B))
	OpInstanceOf // R(A) := R(B) instanceof R(C)

	// Loop operations
	OpForPrep // R(A) -= R(A+2); PC += sBx
	OpForLoop // R(A) += R(A+2); if R(A) <= R(A+1) then PC += sBx; R(A+3) = R(A)

	// Closure operations
	OpClosure // R(A) := closure(KPROTO[Bx])
	OpClose   // close all variables in the stack up to (>=) R(A)

	// Special operations
	OpNop   // no operation
	OpHalt  // halt execution
	OpDebug // debug breakpoint

	OpCodeMax
)

// Instruction represents a 32-bit virtual machine instruction
type Instruction uint32

// Instruction format constants
const (
	OpCodeBits = 6  // 6 bits for opcode (64 opcodes max)
	ABits      = 8  // 8 bits for A operand (256 values)
	BBits      = 9  // 9 bits for B operand (512 values)
	CBits      = 9  // 9 bits for C operand (512 values)
	BxBits     = 18 // 18 bits for Bx operand (262144 values)
	AxBits     = 26 // 26 bits for Ax operand (67108864 values)

	OpCodeMask = (1 << OpCodeBits) - 1
	AMask      = (1 << ABits) - 1
	BMask      = (1 << BBits) - 1
	CMask      = (1 << CBits) - 1
	BxMask     = (1 << BxBits) - 1
	AxMask     = (1 << AxBits) - 1

	MaxA  = AMask
	MaxB  = BMask
	MaxC  = CMask
	MaxBx = BxMask
	MaxAx = AxMask

	// Signed Bx offset
	BxOffset = (1 << (BxBits - 1)) // 131072
)

// Instruction format types
type InstructionFormat int

const (
	FormatABC InstructionFormat = iota // OpCode(8) A(8) B(8) C(8)
	FormatABx                          // OpCode(8) A(8) Bx(16)
	FormatAx                           // OpCode(8) Ax(24)
)

// OpCodeInfo contains metadata about an opcode
type OpCodeInfo struct {
	Name   string
	Format InstructionFormat
	HasA   bool // instruction sets register A
	HasB   bool // instruction uses operand B
	HasC   bool // instruction uses operand C
}

// OpCodeInfos contains information about all opcodes
var OpCodeInfos = [OpCodeMax]OpCodeInfo{
	OpMove:     {"MOVE", FormatABC, true, true, false},
	OpLoadK:    {"LOADK", FormatABx, true, false, false},
	OpLoadNil:  {"LOADNIL", FormatABC, true, false, false},
	OpLoadBool: {"LOADBOOL", FormatABC, true, true, true},
	OpLoadInt:  {"LOADINT", FormatABx, true, false, false},

	OpAdd: {"ADD", FormatABC, true, true, true},
	OpSub: {"SUB", FormatABC, true, true, true},
	OpMul: {"MUL", FormatABC, true, true, true},
	OpDiv: {"DIV", FormatABC, true, true, true},
	OpMod: {"MOD", FormatABC, true, true, true},
	OpPow: {"POW", FormatABC, true, true, true},
	OpNeg: {"NEG", FormatABC, true, true, false},

	OpBitAnd: {"BITAND", FormatABC, true, true, true},
	OpBitOr:  {"BITOR", FormatABC, true, true, true},
	OpBitXor: {"BITXOR", FormatABC, true, true, true},
	OpBitNot: {"BITNOT", FormatABC, true, true, false},
	OpShl:    {"SHL", FormatABC, true, true, true},
	OpShr:    {"SHR", FormatABC, true, true, true},

	OpEq: {"EQ", FormatABC, false, true, true},
	OpNe: {"NE", FormatABC, false, true, true},
	OpLt: {"LT", FormatABC, false, true, true},
	OpLe: {"LE", FormatABC, false, true, true},
	OpGt: {"GT", FormatABC, false, true, true},
	OpGe: {"GE", FormatABC, false, true, true},

	OpNot: {"NOT", FormatABC, true, true, false},
	OpAnd: {"AND", FormatABC, true, true, true},
	OpOr:  {"OR", FormatABC, true, true, true},

	OpJmp:     {"JMP", FormatABx, false, false, false},
	OpTest:    {"TEST", FormatABC, false, true, false},
	OpTestSet: {"TESTSET", FormatABC, true, true, false},

	OpCall:     {"CALL", FormatABC, true, true, true},
	OpTailCall: {"TAILCALL", FormatABC, false, true, true},
	OpReturn:   {"RETURN", FormatABC, false, true, false},

	OpNewTable:  {"NEWTABLE", FormatABC, true, true, true},
	OpGetTable:  {"GETTABLE", FormatABC, true, true, true},
	OpSetTable:  {"SETTABLE", FormatABC, false, true, true},
	OpGetGlobal: {"GETGLOBAL", FormatABx, true, false, false},
	OpSetGlobal: {"SETGLOBAL", FormatABx, false, false, false},
	OpGetUpval:  {"GETUPVAL", FormatABC, true, true, false},
	OpSetUpval:  {"SETUPVAL", FormatABC, false, true, false},

	OpNewArray: {"NEWARRAY", FormatABx, true, false, false},
	OpGetIndex: {"GETINDEX", FormatABC, true, true, true},
	OpSetIndex: {"SETINDEX", FormatABC, false, true, true},
	OpLen:      {"LEN", FormatABC, true, true, false},

	OpConcat: {"CONCAT", FormatABC, true, true, true},

	OpTypeOf:     {"TYPEOF", FormatABC, true, true, false},
	OpInstanceOf: {"INSTANCEOF", FormatABC, true, true, true},

	OpForPrep: {"FORPREP", FormatABx, false, false, false},
	OpForLoop: {"FORLOOP", FormatABx, false, false, false},

	OpClosure: {"CLOSURE", FormatABx, true, false, false},
	OpClose:   {"CLOSE", FormatABC, false, true, false},

	OpNop:   {"NOP", FormatABC, false, false, false},
	OpHalt:  {"HALT", FormatABC, false, false, false},
	OpDebug: {"DEBUG", FormatABC, false, false, false},
}

// CreateABC creates an ABC format instruction
func CreateABC(op OpCode, a, b, c int) Instruction {
	return Instruction(op) |
		Instruction(a&AMask)<<OpCodeBits |
		Instruction(b&BMask)<<(OpCodeBits+ABits) |
		Instruction(c&CMask)<<(OpCodeBits+ABits+BBits)
}

// CreateABx creates an ABx format instruction
func CreateABx(op OpCode, a, bx int) Instruction {
	return Instruction(op) |
		Instruction(a&AMask)<<OpCodeBits |
		Instruction(bx&BxMask)<<(OpCodeBits+ABits)
}

// CreateAx creates an Ax format instruction
func CreateAx(op OpCode, ax int) Instruction {
	return Instruction(op) |
		Instruction(ax&AxMask)<<OpCodeBits
}

// GetOpCode extracts the opcode from an instruction
func (inst Instruction) GetOpCode() OpCode {
	return OpCode(inst & OpCodeMask)
}

// GetA extracts the A operand from an instruction
func (inst Instruction) GetA() int {
	return int((inst >> OpCodeBits) & AMask)
}

// GetB extracts the B operand from an instruction
func (inst Instruction) GetB() int {
	return int((inst >> (OpCodeBits + ABits)) & BMask)
}

// GetC extracts the C operand from an instruction
func (inst Instruction) GetC() int {
	return int((inst >> (OpCodeBits + ABits + BBits)) & CMask)
}

// GetBx extracts the Bx operand from an instruction
func (inst Instruction) GetBx() int {
	return int((inst >> (OpCodeBits + ABits)) & BxMask)
}

// GetSBx extracts the signed Bx operand from an instruction
func (inst Instruction) GetSBx() int {
	return inst.GetBx() - BxOffset
}

// GetAx extracts the Ax operand from an instruction
func (inst Instruction) GetAx() int {
	return int((inst >> OpCodeBits) & AxMask)
}

// String returns a string representation of the instruction
func (inst Instruction) String() string {
	op := inst.GetOpCode()
	if op >= OpCodeMax {
		return fmt.Sprintf("INVALID(%d)", op)
	}

	info := OpCodeInfos[op]
	switch info.Format {
	case FormatABC:
		a, b, c := inst.GetA(), inst.GetB(), inst.GetC()
		if info.HasA && info.HasB && info.HasC {
			return fmt.Sprintf("%-10s R%d, R%d, R%d", info.Name, a, b, c)
		} else if info.HasA && info.HasB {
			return fmt.Sprintf("%-10s R%d, R%d", info.Name, a, b)
		} else if info.HasA {
			return fmt.Sprintf("%-10s R%d", info.Name, a)
		} else if info.HasB && info.HasC {
			return fmt.Sprintf("%-10s R%d, R%d", info.Name, b, c)
		} else {
			return info.Name
		}
	case FormatABx:
		a, bx := inst.GetA(), inst.GetBx()
		if info.HasA {
			return fmt.Sprintf("%-10s R%d, %d", info.Name, a, bx)
		} else {
			return fmt.Sprintf("%-10s %d", info.Name, bx)
		}
	case FormatAx:
		ax := inst.GetAx()
		return fmt.Sprintf("%-10s %d", info.Name, ax)
	default:
		return fmt.Sprintf("UNKNOWN_FORMAT(%s)", info.Name)
	}
}

// IsJump returns true if the instruction is a jump instruction
func (inst Instruction) IsJump() bool {
	op := inst.GetOpCode()
	return op == OpJmp || op == OpTest || op == OpTestSet ||
		op == OpEq || op == OpNe || op == OpLt || op == OpLe ||
		op == OpGt || op == OpGe || op == OpForPrep || op == OpForLoop
}

// IsCall returns true if the instruction is a call instruction
func (inst Instruction) IsCall() bool {
	op := inst.GetOpCode()
	return op == OpCall || op == OpTailCall
}

// IsReturn returns true if the instruction is a return instruction
func (inst Instruction) IsReturn() bool {
	return inst.GetOpCode() == OpReturn
}