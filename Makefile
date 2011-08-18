include $(GOROOT)/src/Make.inc

TARG=sevenzip
CGOFILES=sevenzip.go
CGO_OFILES=7zAlloc.o 7zBuf.o 7zBuf2.o 7zCrc.o 7zCrcOpt.o 7zDec.o 7zIn.o CpuArch.o LzmaDec.o Lzma2Dec.o Bra.o Bra86.o Bcj2.o Ppmd7.o Ppmd7Dec.o 7zFile.o 7zStream.o

include $(GOROOT)/src/Make.pkg
