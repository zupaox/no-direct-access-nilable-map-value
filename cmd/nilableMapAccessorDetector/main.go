package main

import (
	"fmt"
	"go/ast"
	"os"
	"reflect"
	"strings"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/multichecker"
)

var analyzer = &analysis.Analyzer{
	Name: "NilableMapAccessorDetector",
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

	// printNodes(roots, 1)
	unsafeAccesses := detectUnsafeMapAccess(allNodes, pass)
	if len(unsafeAccesses) > 0 {
		printDetectResults(unsafeAccesses)
		os.Exit(1)
	}
	return nil, nil
}

func printNodes(root []*MyNode, level int) {
	for _, node := range root {
		fmt.Printf("%v %v(%v)\n", strings.Repeat("  ", level), reflect.TypeOf(node.AstNode), node.AstNode)
		printNodes(node.Children, level+1)
	}
}

func printDetectResults(results []DetectResult) {
	for _, result := range results {
		fmt.Printf("Detected issue: %v, file: %v, line: %v, line pos: %v\n", result.Issue, result.FileName, result.LineNumber, result.LinePos)
	}
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

type DetectResult struct {
	Issue      string
	FileName   string
	LineNumber int
	LinePos    int
}

func detectUnsafeMapAccess(allNodes []*MyNode, pass *analysis.Pass) []DetectResult {
	var detected []DetectResult
	for _, n := range allNodes {
		indexExpr, ok := n.AstNode.(*ast.IndexExpr)
		if !ok {
			continue
		}

		mapIden, ok := n.Children[0].AstNode.(ast.Expr)
		if !ok {
			// we always expect the first child of a indexExpr to be an identifier
			// if not, we skip it
			fmt.Printf("incorrect first child type for node: %v, type of first child: %v\n", n.AstNode, reflect.TypeOf(n.Children[0].AstNode))
			continue
		}

		childType := pass.TypesInfo.TypeOf(mapIden)
		if !strings.HasPrefix(childType.String(), "map") {
			// both slice access and map access will be categoried as IndexExpr
			// so we only process map accesses
			continue
		}

		sourceFile := pass.Fset.File(n.AstNode.Pos())
		filePath := sourceFile.Name()
		lineNumber := sourceFile.Line(n.AstNode.Pos())
		lineStart := sourceFile.LineStart(lineNumber)
		linePos := n.AstNode.Pos() - lineStart
		assStmt, parentIsAssignment := n.Parent.AstNode.(*ast.AssignStmt)

		indexType := pass.TypesInfo.TypeOf(indexExpr)
		valueIsNilable := strings.HasPrefix(indexType.String(), "func") || strings.HasPrefix(indexType.String(), "(func") || strings.HasPrefix(indexType.String(), "*") || strings.HasPrefix(indexType.String(), "(*")

		if !valueIsNilable {
			continue
		}

		if parentIsAssignment {
			lhs := assStmt.Lhs

			_, firstIsVar := lhs[0].(*ast.Ident)

			if len(lhs) < 2 {
				if firstIsVar {
					detected = append(detected, DetectResult{
						Issue:      "assignment lhs variable length < 2",
						FileName:   filePath,
						LineNumber: lineNumber,
						LinePos:    int(linePos),
					})
				}

				// for assign value to a map like: someMap["abc"] = d
				// we should not flag it
				continue
			}

			secondVar, ok := lhs[1].(*ast.Ident)
			if ok && secondVar.Name == "_" {
				detected = append(detected, DetectResult{
					Issue:      "skipped key existence checking with _",
					FileName:   filePath,
					LineNumber: lineNumber,
					LinePos:    int(linePos),
				})
			}
			continue
		}

		detected = append(detected, DetectResult{
			Issue:      "direct access to nilable map value",
			FileName:   filePath,
			LineNumber: lineNumber,
			LinePos:    int(linePos),
		})
	}

	return detected
}

func main() {
	multichecker.Main(analyzer)
}
