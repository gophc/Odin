package cmd

// ExprInfo holds expression type-checking information
type ExprInfo struct {
	Mode  AddressingMode
	IsLhs bool
	Type  *Type
	Value ExactValue
}

func MakeExprInfo(mode AddressingMode, type_ *Type, value ExactValue, isLhs bool) *ExprInfo {
	return &ExprInfo{
		Mode:  mode,
		Type:  type_,
		Value: value,
		IsLhs: isLhs,
	}
}

type ExprKind int

const (
	ExprKind_Expr ExprKind = iota
	ExprKind_Stmt
)

type StmtFlag int

const (
	StmtFlag_BreakAllowed       StmtFlag = 1 << 0
	StmtFlag_ContinueAllowed    StmtFlag = 1 << 1
	StmtFlag_FallthroughAllowed StmtFlag = 1 << 2
	StmtFlag_TypeSwitch         StmtFlag = 1 << 4
	StmtFlag_CheckScopeDecls    StmtFlag = 1 << 5
)

type BuiltinProcPkg int

const (
	BuiltinProcPkg_builtin    BuiltinProcPkg = iota
	BuiltinProcPkg_intrinsics
	BuiltinProcPkg_COUNT
)

var builtinProcPkgName = [...]string{
	"builtin",
	"intrinsics",
}

type BuiltinProc struct {
	Name          String
	ArgCount      isize
	Variadic      bool
	Kind          ExprKind
	Pkg           BuiltinProcPkg
	Diverging     bool
	IgnoreResults bool
}

type Operand struct {
	Mode      AddressingMode
	Type      *Type
	Value     ExactValue
	Expr      *Ast
	BuiltinID BuiltinProcId
	ProcGroup *Entity
}

type BlockLabel struct {
	Name  String
	Label *Ast
}

type DeferredProcedureKind int

const (
	DeferredProcedure_none         DeferredProcedureKind = iota
	DeferredProcedure_in
	DeferredProcedure_out
	DeferredProcedure_in_out
	DeferredProcedure_in_by_ptr
	DeferredProcedure_out_by_ptr
	DeferredProcedure_in_out_by_ptr
)

type DeferredProcedure struct {
	Kind   DeferredProcedureKind
	Entity *Entity
}

type InstrumentationFlag int32

const (
	Instrumentation_Enabled  InstrumentationFlag = -1
	Instrumentation_Default  InstrumentationFlag = 0
	Instrumentation_Disabled InstrumentationFlag = +1
)

type AttributeContext struct {
	LinkName                 String
	LinkPrefix               String
	LinkSuffix               String
	LinkSection              String
	Linkage                  String
	InitExprListCount        isize
	ThreadLocalModel         String
	DeprecatedMessage        String
	WarningMessage           String
	DeferredProcedure        DeferredProcedure
	IsExport                 bool
	IsStatic                 bool
	RequireResults           bool
	RequireDeclaration       bool
	HasDisabledProc          bool
	DisabledProc             bool
	Test                     bool
	Init                     bool
	Fini                     bool
	SetCold                  bool
	EntryPointOnly           bool
	InstrumentationEnter     bool
	InstrumentationExit      bool
	NoSanitizeAddress        bool
	NoSanitizeMemory         bool
	NoSanitizeThread         bool
	Rodata                   bool
	IgnoreDuplicates         bool
	OptimizationMode         uint32
	ForeignImportPriorityIndex int64
	ExtraLinkerFlags         String
	NoInstrumentation        InstrumentationFlag
	ObjcClass                String
	ObjcName                 String
	ObjcSelector             String
	ObjcType                 *Type
	ObjcSuperclass           *Type
	ObjcIvar                 *Type
	ObjcContextProvider      *Entity
	ObjcIsClassMethod        bool
	ObjcIsImplementation     bool
	ObjcIsDisabledImplement  bool
	RequireTargetFeature     String
	EnableTargetFeature      String
	RaddbgTypeView           bool
	RaddbgTypeViewString     String
}

func MakeAttributeContext(linkPrefix, linkSuffix String) AttributeContext {
	return AttributeContext{
		LinkPrefix: linkPrefix,
		LinkSuffix: linkSuffix,
	}
}

type DeclAttributeProc func(c *CheckerContext, elem *Ast, name String, value *Ast, ac *AttributeContext) bool

func CheckDeclAttributes(c *CheckerContext, attributes []*Ast, proc DeclAttributeProc, ac *AttributeContext) { check_decl_attributes(c, attributes, proc, ac) }
func CheckArityMatch(c *CheckerContext, vd *AstValueDecl, isGlobal bool) bool            { return check_arity_match(c, vd, isGlobal) }
func CheckCollectEntitiesFromWhenStmt(c *CheckerContext, ws *Ast)                         { check_collect_entities_from_when_stmt(c, ws) }
func CheckDelayedFileImportEntity(c *CheckerContext, decl *Ast)                          {}
func NewCheckerTypePath() *CheckerTypePath                                                { return &CheckerTypePath{} }
func DestroyCheckerTypePath(tp *CheckerTypePath)                                         {}
func CheckTypePathPush(c *CheckerContext, e *Entity)                                     {}
func CheckTypePathPop(c *CheckerContext) *Entity                                         { return nil }
func InitCoreContext(c *Checker)                                                         {}
func InitMemAllocator(c *Checker)                                                        {}
func AddUntypedExpressions(cinfo *CheckerInfo, untyped *UntypedExprInfoMap)              {}
func EnsurePolymorphicRecordEntityHasGenTypes(ctx *CheckerContext, originalType *Type) *GenTypesData { return nil }
func InitMapInternalTypes(type_ *Type)                                                   {}
