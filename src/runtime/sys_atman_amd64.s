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
	ADDQ	$8, DI	// ELF wants to use -8(FS)

	MOVQ	DI, SI
	MOVQ	$0x1002, DI	// ARCH_SET_FS
	MOVQ	$158, AX	// arch_prctl
	SYSCALL
	CMPQ	AX, $0xfffffffffffff001
	JLS	2(PC)
	MOVL	$0xf1, 0xf1  // crash
	RET
