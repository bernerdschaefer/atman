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

	// set _atman_phys_to_machine_mapping
	MOVQ	runtime·_atman_start_info+0(SB), AX
	ADDQ	$104, (AX)
	MOVQ	AX, runtime·_atman_phys_to_machine_mapping(SB)

	// map gdt page to machine frame number
	MOVQ	runtime·_atman_gdt_page(SB), AX
	SUBQ	0x401000, AX // address of text
	MOVQ	0x12, CX
	SHRQ	CX, AX     // shift-right 12
	MOVQ	AX, CX
	MOVQ	0x8, AX
	MULQ	CX         // multiply by 8 for offset
	MOVQ	CX, AX
	MOVQ	runtime·_atman_phys_to_machine_mapping(SB), CX
	ADDQ	AX, (CX)   // move to pfn offset

	MOVQ	CX, DI // mfn
	MOVQ	$1, SI // length 1
	MOVQ	$runtime·_atman_hypercall_page+0x40(SB), AX // set_gdt
	// callq *%rax
	BYTE $0xFF; BYTE $0xd0
	MOVQ	AX, (SP)

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
