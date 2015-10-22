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
TEXT runtime·settls(SB),NOSPLIT,$32
	MOVQ	DI, SI	// arg2 = tls base
	MOVQ	$0, DI	// arg1 = fs (0)
	MOVQ	$runtime·_atman_hypercall_page+0x320(SB), AX
	// callq *%rax
	BYTE $0xFF; BYTE $0xd0
	MOVQ	AX, (SP)
	RET
