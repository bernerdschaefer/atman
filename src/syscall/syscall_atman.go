package syscall

var EINVAL = errorString("bad arg in system call")

type Timespec struct {
	Sec  int64
	Nsec int32
}

type Timeval struct {
	Sec  int64
	Usec int32
}

type errorString string

func (s errorString) Error() string { return string(s) }
