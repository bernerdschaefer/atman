package runtime

import "unsafe"

// Atman implements a round-robin cooperative scheduler,
// heavily inspired by libtask.

var (
	taskid = 1
	taskn  = 1

	// taskcurrent is initialized as an empty Task with ID 0.
	// The first time a context switch is requested,
	// its context will be filled in and it will become
	// a normal Task.
	taskcurrent = &Task{ID: 0}

	taskrunqueue TaskList
)

type Task struct {
	ID    int
	State [256]byte

	Context Context

	Ready bool

	Next, Prev *Task
}

// taskcreate spawns a new task,
// executing the argumentless function fn
// on the provided stack stk
func taskcreate(mp, g0, fn, stk unsafe.Pointer) {
	// write fn to stack
	*(*uintptr)(stk) = uintptr(fn)
	stk = add(stk, ptrSize)

	// write mp to stack
	*(*uintptr)(stk) = uintptr(mp)
	stk = add(stk, ptrSize)

	// write g0 to stack
	*(*uintptr)(stk) = uintptr(g0)
	stk = add(stk, ptrSize)

	t := &Task{
		ID: taskid,
		Context: Context{
			rsp: uintptr(stk),
			rip: funcPC(taskstart),
		},
	}

	taskid++
	taskn++

	taskready(t)
}

//go:nosplit
func taskstart(fn, mp, gp unsafe.Pointer)

func taskready(t *Task) {
	t.Ready = true
	taskrunqueue.Add(t)
}

func taskyield() {
	taskready(taskcurrent)
	taskswitch()
}

func taskswitch() {
	t := taskrunqueue.Head
	taskrunqueue.Remove(t)
	t.Ready = false

	contextswitch(&taskcurrent.Context, &t.Context)

	// anything else?
}

func taskexit() {
	println("taskexit()")
}

type TaskList struct {
	Head, Tail *Task
}

func (l *TaskList) Add(t *Task) {
	if l.Tail != nil {
		l.Tail.Next = t
		t.Prev = l.Tail
	} else {
		l.Head = t
		t.Prev = nil
	}

	l.Tail = t
	t.Next = nil
}

func (l *TaskList) Remove(t *Task) {
	if t.Prev != nil {
		t.Prev.Next = t.Next
	} else {
		l.Head = t.Next
	}

	if t.Next != nil {
		t.Next.Prev = t.Prev
	} else {
		l.Tail = t.Prev
	}
}

// Context describes the state of a task
// for saving or restoring a task's execution context.
type Context cpuRegisters

func contextswitch(from, to *Context) {
	contextsave(from)
	contextload(to)
}

func contextsave(*Context)
func contextload(*Context)

// cpuRegisters describes the state of a CPU
// on entry of a trap or interrupt.
type cpuRegisters struct {
	r15    uintptr
	r14    uintptr
	r13    uintptr
	r12    uintptr
	rbp    uintptr
	rbx    uintptr
	r11    uintptr
	r10    uintptr
	r9     uintptr
	r8     uintptr
	rax    uintptr
	rcx    uintptr
	rdx    uintptr
	rsi    uintptr
	rdi    uintptr
	code   uintptr // syscall number, error code, or IRQ number
	rip    uintptr
	cs     uintptr
	rflags uintptr
	rsp    uintptr
	ss     uintptr
	es     uintptr
	ds     uintptr
	fs     uintptr
	gs     uintptr
}
