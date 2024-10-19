/*
	GlowDash - Smart Home Web Dashboard

	(C) 2024 Péter Deák (hyper80@gmail.com)
	License: GPLv2
*/

package main

type IntStack []int

func NewIntStack() IntStack {
	return []int{}
}

func (s *IntStack) Clear() {
	*s = []int{}
}

func (stack *IntStack) Push(v int) {
	*stack = append(*stack, v)
}

func (s *IntStack) Pop() int {
	l := len(*s)
	if l == 0 {
		return 0
	}
	val := (*s)[l-1]
	*s = (*s)[:l-1]
	return val
}

func (s *IntStack) PopDef(def int) int {
	if s.IsEmpty() {
		return def
	}
	return s.Pop()
}

func (s IntStack) IsEmpty() bool {
	if len(s) == 0 {
		return true
	}
	return false
}

func (s IntStack) Top() int {
	l := len(s)
	if l == 0 {
		return 0
	}
	return s[l-1]
}

func (s IntStack) TopDef(def int) int {
	if s.IsEmpty() {
		return def
	}
	return s.Top()
}
