package main

import (
	"fmt"
	"log"
	"strconv"
)

type Node struct {
	Value string
	Left  *Node
	Right *Node
}

func (n *Node) isNumber() bool {
	_, err := strconv.ParseInt(n.Value, 10, 64)

	return err == nil
}

func (n *Node) returnNumberValue() int64 {
	number, err := strconv.ParseInt(n.Value, 10, 64)

	if err != nil {
		log.Fatal(err)
	}
	return number

}

func (n *Node) performOperation(left, right *Node) int {
	switch n.Value {
	case "*":
		return left.evaluate() * right.evaluate()
	case "/":
		return left.evaluate() / right.evaluate()
	case "+":
		return left.evaluate() + right.evaluate()
	case "-":
		return left.evaluate() - right.evaluate()
	default:
		panic("Unknown operation")
	}
}

func (n *Node) evaluate() int {
	if n == nil {
		return 0
	}
	if n.isNumber() {
		return int(n.returnNumberValue())
	}
	return n.performOperation(n.Left, n.Right)
}

func main() {
	expression := Node{Value: "+", Left: &Node{Value: "*", Left: &Node{Value: "6"}, Right: &Node{Value: "6"}}, Right: &Node{Value: "/", Left: &Node{Value: "81"}, Right: &Node{Value: "3"}}}

	fmt.Printf("evaluated expression equals to: %d\n", expression.evaluate())

}
