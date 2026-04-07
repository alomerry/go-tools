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
		// 情况 1: 基本字面量 (如 "abc", 123) -> 不通过
		if val, success := arg.(*ast.BasicLit); success {
			pass.Reportf(callExpr.Pos(), "[%s.%s(%s] 不合法，必须导入常量", receiver, methodName, val.Value)
			return
		}
		// 情况 2: 标识符 (如 Admin, myVar)
		if _, success := arg.(*ast.Ident); success {
			pass.Reportf(callExpr.Pos(), "[%s.%s] 参数不合法，必须导入 [%s] 包常量", receiver, methodName, argIdx.consPkg)
			return
		}
		// 情况 3: Selector 表达式 (如 common.PIPELINE_NAME)
		if selExpr, success := arg.(*ast.SelectorExpr); success {
			obj := pass.TypesInfo.Uses[selExpr.Sel]
			if obj == nil {
				pass.Reportf(callExpr.Pos(), "类型不存在")
				return
			}

			if _, isConst := obj.(*types.Const); !isConst {
				pass.Reportf(callExpr.Pos(), "必须常量")
				return
			}

			if obj.Pkg() == nil {
				pass.Reportf(callExpr.Pos(), "包路径不存在")
				return
			}

			if obj.Pkg().Path() != argIdx.consPkg {
				pass.Reportf(callExpr.Pos(), "[%s.%s] 参数 %d 不合法，目前导入了包 [%s] 常量，要求必须导入 [%s] 包常量", receiver, methodName, argIdx.argIdx, obj.Pkg().Path(), argIdx.consPkg)
				return
			}
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
