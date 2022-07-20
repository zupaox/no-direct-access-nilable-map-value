package main

import (
	"fmt"
	"go/ast"
	"strings"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/multichecker"
)

var analyzer = &analysis.Analyzer{
	Name: "Test",
	Doc:  "reports warning when code trys to access map value that is nilable without using assignment",
	Run:  runAnalysis,
}

type MyNode struct {
	AstNode  ast.Node
	Children []*MyNode
	Parent   *MyNode
	File     *ast.File
}

func runAnalysis(pass *analysis.Pass) (interface{}, error) {
	allNodes := []*MyNode{}
	roots := []*MyNode{}
	for _, file := range pass.Files {
		ast.Inspect(file, func(n ast.Node) bool {
			if n != nil {
				n := MyNode{
					AstNode:  n,
					Children: []*MyNode{},
					Parent:   nil,
					File:     file,
				}
				found := addParent(&n, allNodes)
				if !found {
					roots = append(roots, &n)
				}
				allNodes = append(allNodes, &n)
			}
			return true
		})
	}

	detectUnsafeMapAccess(allNodes, pass)
	return nil, nil
}

// because go/analysis always walk the nodes depth first
// so the last node we found that contains this node, is always the direct parent of it
func addParent(n *MyNode, allNodes []*MyNode) bool {
	pos := n.AstNode.Pos()
	end := n.AstNode.End()

	for i := len(allNodes) - 1; i >= 0; i-- {
		p := allNodes[i]
		if p.AstNode.Pos() <= pos && p.AstNode.End() >= end {
			p.Children = append(p.Children, n)
			n.Parent = p
			return true
		}
	}

	return false
}

func printMyNodes(level int, roots []*MyNode, formatter func(n *MyNode) string) {
	for _, n := range roots {
		fmt.Printf("%v %v\n", strings.Repeat("  ", level), formatter(n))
		printMyNodes(level+1, n.Children, formatter)
	}
}

func detectUnsafeMapAccess(allNodes []*MyNode, pass *analysis.Pass) {
	for _, n := range allNodes {
		indexExpr, ok := n.AstNode.(*ast.IndexExpr)
		if !ok {
			continue
		}

		sourceFile := pass.Fset.File(n.AstNode.Pos())
		filePath := sourceFile.Name()
		lineNumber := sourceFile.Line(n.AstNode.Pos())
		lineStart := sourceFile.LineStart(lineNumber)
		linePos := n.AstNode.Pos() - lineStart
		assStmt, parentIsAssignment := n.Parent.AstNode.(*ast.AssignStmt)

		indexType := pass.TypesInfo.TypeOf(indexExpr)
		valueIsNilable := strings.HasPrefix(indexType.String(), "func") || strings.HasPrefix(indexType.String(), "*command-line-arguments")

		if parentIsAssignment && valueIsNilable {
			lhs := assStmt.Lhs
			if len(lhs) < 2 {
				fmt.Printf("warning (%v: %v %v): Please ensure key existed first (len < 2).\n", filePath, lineNumber, linePos)
				continue
			}
			variable, ok := lhs[1].(*ast.Ident)
			if ok && variable.Name == "_" {
				fmt.Printf("warning (%v: %v %v): Please ensure key existed (underscore).\n", filePath, lineNumber, linePos)
			}
			continue
		}

		if valueIsNilable {
			fmt.Printf("warning (%v: %v %v): Please ensure key existed (direct access, %v).\n", filePath, lineNumber, linePos, indexType.String())
		}
	}
}

func main() {
	multichecker.Main(analyzer)
}
