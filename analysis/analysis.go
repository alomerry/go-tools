package analysis

import (
	"go/ast"
	"go/types"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

var CustomAnalyzer = &analysis.Analyzer{
	Name:     "custom_vet",
	Doc:      "custom_vet",
	Requires: []*analysis.Analyzer{inspect.Analyzer},
	Run: func(pass *analysis.Pass) (any, error) {
		// 从依赖中获取 inspect 的结果
		inspectResult := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)

		// 我们只关心函数调用表达式 (CallExpr)
		nodeFilter := []ast.Node{
			(*ast.CallExpr)(nil),
		}

		// 遍历过滤后的节点
		inspectResult.Preorder(nodeFilter, check(pass))
		return nil, nil
	},
}

func check(pass *analysis.Pass) func(n ast.Node) {
	return func(n ast.Node) {
		checkGenUseCons(pass, n)
	}
}

var (
	consValidateConfigs = map[string]map[string]map[string][]consArg{
		"github.com/alomerry/go-tools/components/redis": {
			"Generator": {
				"GenKey": []consArg{{0, "github.com/alomerry/go-tools/static/cons"}},
			},
		},
	}
)

type consArg struct {
	argIdx  int
	consPkg string
}

func checkGenUseCons(pass *analysis.Pass, n ast.Node) {
	callExpr := n.(*ast.CallExpr)
	// 方法调用通常表现为 SelectorExpr，例如 u.SetRole
	selExpr, ok := callExpr.Fun.(*ast.SelectorExpr)
	if !ok {
		return
	}

	// 检查接收者（X）的类型是否是目标结构体
	tv, ok := pass.TypesInfo.Types[selExpr.X]
	if !ok {
		return
	}

	xType := tv.Type
	if xType == nil {
		return
	}

	// 检查接收者（X）是否来自目标包
	pkgConfig, exists := consValidateConfigs[getTypePkgPath(xType)]
	if !exists {
		return
	}

	// 这里可以复用你原来的 getTypeName 逻辑
	receiver := getTypeName(xType)
	methodConfig, exists := pkgConfig[receiver]
	if !exists {
		return
	}

	// 检查方法名是否匹配
	methodName := selExpr.Sel.Name
	argConfig, exists := methodConfig[methodName]
	if !exists || len(argConfig) == 0 {
		return
	}

	if len(callExpr.Args) == 0 {
		pass.Reportf(callExpr.Pos(), "%s.%s 参数异常，数量为 0", receiver, methodName)
		return
	}

	for _, argIdx := range argConfig {
		arg := callExpr.Args[argIdx.argIdx]

		// 1. 拦截基本字面量 (如 "abc", 123) -> 不通过
		if val, success := arg.(*ast.BasicLit); success {
			pass.Reportf(callExpr.Pos(), "[%s.%s(%s] 不合法，必须是变量或导入 [%s] 包常量", receiver, methodName, val.Value, argIdx.consPkg)
			return
		}

		// 2. 提取参数对应的对象 (Object)
		var obj types.Object
		if ident, ok := arg.(*ast.Ident); ok {
			// 情况: 单纯的标识符 (如 myVar, req, Admin)
			obj = pass.TypesInfo.Uses[ident]
		} else if selExpr, ok := arg.(*ast.SelectorExpr); ok {
			// 情况: Selector 表达式 (如 common.PIPELINE_NAME, req.Name)
			obj = pass.TypesInfo.Uses[selExpr.Sel]
		}

		if obj == nil {
			pass.Reportf(callExpr.Pos(), "[%s.%s] 参数形式不合法，必须是变量或导入 [%s] 包常量", receiver, methodName, argIdx.consPkg)
			return
		}

		// 3. 判断对象是变量还是常量
		switch objType := obj.(type) {
		case *types.Var:
			// 这是一个变量 (包括函数入参、局部变量、结构体字段等)
			// 允许通过，无需做额外处理
		case *types.Const:
			// 这是一个常量，进行包路径校验
			if objType.Pkg() == nil {
				pass.Reportf(callExpr.Pos(), "包路径不存在")
				return
			}

			if objType.Pkg().Path() != argIdx.consPkg {
				pass.Reportf(callExpr.Pos(), "[%s.%s] 参数 %d 不合法，目前导入了包 [%s] 常量，要求必须导入 [%s] 包常量", receiver, methodName, argIdx.argIdx, objType.Pkg().Path(), argIdx.consPkg)
				return
			}
		default:
			// 拦截其他的类型（例如函数等）
			pass.Reportf(callExpr.Pos(), "[%s.%s] 参数类型不合法，必须是变量或常量", receiver, methodName)
			return
		}
	}
}

// getTypeName 从类型中提取结构体名称
// 处理 *Struct 和 Struct 的情况
func getTypeName(t types.Type) string {
	// 如果是指针，取指向的类型
	if ptr, ok := t.(*types.Pointer); ok {
		t = ptr.Elem()
	}

	// 如果是命名类型，返回其名称
	if named, ok := t.(*types.Named); ok {
		return named.Obj().Name()
	}

	return ""
}

// 获取类型所在的包路径
func getTypePkgPath(t types.Type) string {
	if ptr, ok := t.(*types.Pointer); ok {
		t = ptr.Elem()
	}

	if named, ok := t.(*types.Named); ok {
		if pkg := named.Obj().Pkg(); pkg != nil {
			return pkg.Path()
		}
	}

	return ""
}
