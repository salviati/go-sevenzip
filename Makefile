include $(GOROOT)/src/Make.inc

TARG=sevenzip
CGOFILES=sevenzip.go
CGO_OFILES=7zAlloc.o 7zBuf.o 7zBuf2.o 7zDec.o 7zIn.o CpuArch.o LzmaDec.o Lzma2Dec.o Bra.o Bra86.o Bcj2.o Ppmd7.o Ppmd7Dec.o 7zFile.o 7zStream.o
USEASM=1

ifeq ($(USEASM), 1)
	ifeq ($(O), 6)
		ASM=yasm -f elf -m amd64
		CGO_OFILES+= 7zCrcT8U.$(O).o 7zCrcT8.o
	else ifeq ($(O), 8)
		ASM=nasm -f elf
		CGO_OFILES+= 7zCrcT8U.$(O).o 7zCrcT8.o
	else
		CGO_OFILES+=7zCrc.o 7zCrcOpt.o
	endif
else
	CGO_OFILES+=7zCrc.o 7zCrcOpt.o
endif

include $(GOROOT)/src/Make.pkg

7zCrcT8U.$(O).o: 7zCrcT8U.$(O).asm
	$(ASM) -o $@ $<
