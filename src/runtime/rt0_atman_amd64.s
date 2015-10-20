TEXT _rt0_amd64_atman(SB),NOSPLIT,$-8
	CLD
	MOVQ	$runtime·_atman_stack+0x8000(SB), SP
	ANDQ	$(~(0x1000-1)), SP

	MOVQ	SI, runtime·_atman_start_info+0(SB)

	MOVQ	$0, DI // CONSOLEIO_write
	MOVQ	$6, SI // strlen(8)
	MOVQ	$runtime·_atman_hello(SB), DX
	MOVQ	$runtime·_atman_hypercall_page+0x240(SB), AX
	// callq *%rax
	BYTE $0xFF; BYTE $0xd0
	MOVQ	AX, (SP)

	MOVQ	runtime·_atman_start_info+0(SB), AX
	ADDQ	$104, (AX)
	MOVQ	(AX), DX
	MOVQ	DX, runtime·_atman_phys_to_machine_mapping+0(SB)

	MOVQ	$0, SI // argv
	MOVQ	$0, DI // argc
	MOVQ	$main(SB), AX
	JMP	AX

TEXT main(SB),NOSPLIT,$-8
	MOVQ	$runtime·rt0_go(SB), AX
	JMP	AX

DATA runtime·isatman(SB)/4, $1
GLOBL runtime·isatman(SB), NOPTR, $4

DATA runtime·hello(SB)/8, $"hello"
GLOBL runtime·hello(SB), NOPTR, $8
