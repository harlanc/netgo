package lib

import "errors"

//Stack can save any types
type Stack []interface{}

//Length get the length of Stack
func (stack *Stack) Length() int {
	return len(*stack)
}

//IsEmpty judge if the stack is empty
func (stack *Stack) IsEmpty() bool {
	return len(*stack) == 0
}

//Cap get the cap of stack
func (stack *Stack) Cap() int {
	return cap(*stack)
}

//Push push a element to Stack
func (stack *Stack) Push(value interface{}) {
	*stack = append(*stack, value)
}

//Top get the top element of the Stack
func (stack *Stack) Top() (interface{}, error) {
	if len(*stack) == 0 {
		return nil, errors.New("Out of index, len is 0")
	}
	return (*stack)[len(*stack)-1], nil
}

//Pop pop a element from Stack
func (stack *Stack) Pop() (interface{}, error) {
	theStack := *stack
	if len(theStack) == 0 {
		return nil, errors.New("Out of index, len is 0")
	}
	value := theStack[len(theStack)-1]
	*stack = theStack[:len(theStack)-1]
	return value, nil
}
