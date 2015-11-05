#define _PAGE_ROUND_UP(REGISTER) \
	ADDQ	$0x0000000000000fff, REGISTER	\
	ANDQ	$0xfffffffffffff000, REGISTER

#define CALL_RBX \
	BYTE $0xff; BYTE $0xd3	// callq *%rbx

#define CRASH \
	ANDQ	$0xdeadbeef, 0xdeadbeef

#define CRASH_ON_NONZERO \
	CMPQ	AX, $-1		\
	JNE	2(PC)		\
	CRASH

#define _HYPERCALL(OFFSET) \
	MOVQ	$runtime·_atman_hypercall_page(SB), BX	\
	_PAGE_ROUND_UP(BX)				\
	ADDQ	OFFSET, BX				\
	CALL_RBX                                        \
	CRASH_ON_NONZERO

#define _HYPERVISOR_console_io(OP, SIZE, DATA_PTR) \
	MOVQ	OP, DI		\
	MOVQ	SIZE, SI	\
	MOVQ	DATA_PTR, DX	\
	_HYPERCALL($0x240)

#define _HYPERVISOR_set_segment_base(WHICH, BASE) \
	MOVQ	WHICH,	DI				\
	MOVQ	BASE, SI				\
	_HYPERCALL($0x320)

TEXT _rt0_amd64_atman(SB),NOSPLIT,$-8
	CLD
	MOVQ	$runtime·_atman_stack+0x8000(SB), SP

	MOVQ	SI, runtime·_atman_start_info(SB)

	MOVQ	$main(SB), AX
	JMP	AX

TEXT main(SB),NOSPLIT,$-8
	MOVQ	$runtime·rt0_go(SB), AX
	JMP	AX

// set tls base to DI
TEXT runtime·settls(SB),NOSPLIT,$32
	MOVQ	DI, AX
	_HYPERVISOR_set_segment_base($0, AX)
	RET

TEXT ·HYPERVISOR_console_io(SB),$24-0
	_HYPERVISOR_console_io(op+0(FP), size+8(FP), data+16(FP))
	RET
