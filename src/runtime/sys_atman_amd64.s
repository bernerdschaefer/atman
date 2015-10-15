TEXT runtime·exit(SB),NOSPLIT,$0-4
	MOVL	code+0(FP), DI
	MOVL	$231, AX	// exitgroup - force all os threads to exit
	SYSCALL
	RET

TEXT runtime·usleep(SB),NOSPLIT,$16
	RET

TEXT runtime·nanotime(SB),NOSPLIT,$16
	RET

TEXT runtime·write(SB),NOSPLIT,$0-28
	RET

// func now() (sec int64, nsec int32)
TEXT time·now(SB),NOSPLIT,$16
	RET

// set tls base to DI
TEXT runtime·settls(SB),NOSPLIT,$0
	RET
