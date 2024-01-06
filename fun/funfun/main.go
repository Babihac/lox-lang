package main

import (
	"fmt"
)

type Node struct {
	value int
	next  *Node
}

type LinkedList struct {
	head *Node
	tail *Node
}

func (l *LinkedList) append(node *Node) {
	if l.head == nil {
		l.head = node
		l.tail = l.head
	} else {
		l.tail.next = node
		l.tail = node
	}
}

func (l *LinkedList) hasCycle() bool {
	turtle := l.head
	haare := l.head.next

	for {
		if turtle == nil || haare == nil {
			return false
		}

		if turtle == haare {
			return true
		}

		turtle = turtle.next
		haare = haare.next.next

	}

}

func main() {
	linkedList := LinkedList{}

	nodeA := Node{value: 358}
	nodeB := Node{value: 34}
	nodeC := Node{value: 12}

	nodeA.next = &nodeB
	nodeB.next = &nodeC
	nodeC.next = &nodeA

	linkedList.append(&Node{value: 1})
	linkedList.append(&Node{value: 2})
	linkedList.append(&Node{value: 3})
	linkedList.append(&Node{value: 4})
	linkedList.append(&Node{value: 5})
	linkedList.append(&Node{value: 6})
	linkedList.append(&Node{value: 7})
	linkedList.append(&nodeA)

	fmt.Printf("hasCycle returned: %t\n", linkedList.hasCycle())

	funs := []int{1, 2, 3}

	done := make(chan bool)

	for _, v := range funs {
		fmt.Println(v)

		go func(val int) {
			fmt.Printf("From GO routine: %d\n", val)
			done <- true
		}(v)
	}

	for range funs {
		<-done
	}

}
