package main

import (
	"fmt"
	"go/ast"
	"strings"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/singlechecker"
)

var analyzer = &analysis.Analyzer{
	Name: "no-direct-access=nilable-map-value",
	Doc:  "reports warning when code trys to access map value that is nilable without using assignment",
	Run:  runAnalysis,
}

type MyNode struct {
	AstNode  ast.Node
	Children []*MyNode
	Parent   *MyNode
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

	/*
		printMyNodes(1, roots, func(n *MyNode) string {
			indexExpr, ok := n.AstNode.(*ast.IndexExpr)
			if !ok {
				return "-"
			}
			return fmt.Sprintf("index express, exp: %v, parent node type: %v, type: %v", indexExpr.X, reflect.TypeOf(n.Parent.AstNode), pass.TypesInfo.TypeOf(indexExpr))
		})
	*/
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

		_, parentIsAssignment := n.Parent.AstNode.(*ast.AssignStmt)
		if parentIsAssignment {
			continue
		}

		indexType := pass.TypesInfo.TypeOf(indexExpr)
		if strings.HasPrefix(indexType.String(), "func") || strings.HasPrefix(indexType.String(), "*command-line-arguments") {
			fmt.Printf("warning (%v): Unsafe direct access to map with func value or pointer value, please use assignment style to access the value and check key existence.\n", n.AstNode)
		}
	}
}

func main() {
	singlechecker.Main(analyzer)
}
