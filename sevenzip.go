/*
   Copyright (c) Utkan Güngördü <utkan@freeconsole.org>

   This program is free software; you can redistribute it and/or modify
   it under the terms of the Lesser GNU General Public License as
   published by the Free Software Foundation; either version 3 or
   (at your option) any later version.

   This program is distributed in the hope that it will be useful,
   but WITHOUT ANY WARRANTY; without even the implied warranty of
   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
   Lesser GNU General Public License for more details


   You should have received a copy of the Lesser GNU General Public
   License along with this program; if not, write to the
   Free Software Foundation, Inc.,
   51 Franklin Street, Fifth Floor, Boston, MA  02110-1301, USA.
*/

package sevenzip
// Package sevenzip implements access to 7-zip archives (wraps C interface of LZMA SDK).

// #include "7z.h"
// #include "7zAlloc.h"
// #include "7zCrc.h"
// #include "7zFile.h"
// #include "7zVersion.h"
//
//void *_szAlloc = SzAlloc;
//void *_szFree = SzFree;
//void *_szAllocTemp = SzAllocTemp;
//void *_szFreeTemp = SzFreeTemp;
//
//static Byte kUtf8Limits[5] = { 0xC0, 0xE0, 0xF0, 0xF8, 0xFC };
//
//static Bool Utf16_To_Utf8(Byte *dest, size_t *destLen, const UInt16 *src, size_t srcLen)
// {
//   size_t destPos = 0, srcPos = 0;
//   for (;;)
//   {
//     unsigned numAdds;
//     UInt32 value;
//     if (srcPos == srcLen)
//     {
//       *destLen = destPos;
//       return True;
//     }
//     value = src[srcPos++];
//     if (value < 0x80)
//     {
//       if (dest)
//         dest[destPos] = (char)value;
//       destPos++;
//       continue;
//     }
//     if (value >= 0xD800 && value < 0xE000)
//     {
//       UInt32 c2;
//       if (value >= 0xDC00 || srcPos == srcLen)
//         break;
//       c2 = src[srcPos++];
//       if (c2 < 0xDC00 || c2 >= 0xE000)
//         break;
//       value = (((value - 0xD800) << 10) | (c2 - 0xDC00)) + 0x10000;
//     }
//     for (numAdds = 1; numAdds < 5; numAdds++)
//       if (value < (((UInt32)1) << (numAdds * 5 + 6)))
//         break;
//     if (dest)
//       dest[destPos] = (char)(kUtf8Limits[numAdds - 1] + (value >> (6 * numAdds)));
//     destPos++;
//     do
//     {
//       numAdds--;
//       if (dest)
//         dest[destPos] = (char)(0x80 + ((value >> (6 * numAdds)) & 0x3F));
//       destPos++;
//     }
//     while (numAdds != 0);
//   }
//   *destLen = destPos;
//   return False;
// }
//
//const void * _null = (void*)0;
//
//void _IAlloc_Free(ISzAlloc *p, void *a) { IAlloc_Free(p,a); }
//Byte *_ByteArrayIndex(Byte *p, size_t i) { return &p[i]; }
//CSzFileItem *_CSzFileItemArrayIndex(CSzFileItem *p, size_t i) { return &p[i]; }
// SRes _SZ_OK = SZ_OK;
import "C"

import (
	"unsafe"
	"os"
	"sync"
	"reflect"
	"bytes"
)

type SevenZip struct {
	allocImp, allocTempImp C.ISzAlloc
	archiveStream          C.CFileInStream
	lookStream             C.CLookToRead
	db                     C.CSzArEx

	blockIndex    C.UInt32
	outBufferSize C.size_t
	outBuffer     *C.Byte

	l sync.Mutex

	File []*File
}

type NtfsFileTime struct {
	Low  uint32
	High uint32
}

type FileHeader struct {
	MTime         NtfsFileTime
	Size          uint64
	Crc           uint32
	Attrib        uint32
	HasStream     byte
	IsDir         byte
	IsAnti        byte
	CrcDefined    byte
	MTimeDefined  byte
	AttribDefined byte
}

type File struct {
	*FileHeader
	Name  string
	Index int
	z     *SevenZip
	buf   []byte
}

// Returns the buffered contents of an opened file.
// This function won't make "unnecessary" copies of the underlying data.
func (f *File) ReadAll() []byte {
	return f.buf
}

// This function will call Extract.
func (f *File) Open() (*bytes.Buffer, os.Error) {
	buf, err := f.z.Extract(f.Index)
	if err != nil {
		return nil, err
	}
	f.buf = buf
	return bytes.NewBuffer(buf), err
}

func (f *File) Close() {
	
}

// This function will call ExtractUnsafe.
func (f *File) OpenUnsafe() (*bytes.Buffer, os.Error) {
	buf, err := f.z.ExtractUnsafe(f.Index)
	if err != nil {
		return nil, err
	}
	f.buf = buf
	return bytes.NewBuffer(buf), err
}

func init() {
	C.CrcGenerateTable()
}

func (z *SevenZip) name(i int) string {
	length := C.SzArEx_GetFileNameUtf16(&z.db, C.size_t(i), (*C.UInt16)(C._null))
	temp := C.SzAlloc(C._null, length*2)
	if temp == C._null {
		panic("sevenzip: out of memory")
	}
	defer C.SzFree(C._null, temp)

	C.SzArEx_GetFileNameUtf16(&z.db, C.size_t(i), (*C.UInt16)(temp))

	var destLen C.size_t
	C.Utf16_To_Utf8((*C.Byte)(C._null), &destLen, (*C.UInt16)(temp), length)
	destLen++
	temp2 := C.SzAlloc(C._null, destLen)
	if temp2 == C._null {
		panic("sevenzip: out of memory")
	}
	defer C.SzFree(C._null, temp2)

	C.Utf16_To_Utf8((*C.Byte)(temp2), &destLen, (*C.UInt16)(temp), length)

	name := C.GoString((*C.char)(temp2))
	return name
}

func (z *SevenZip) flushBuffer() {
	if z.outBuffer != (*C.Byte)(C._null) {
		C._IAlloc_Free(&z.allocImp, unsafe.Pointer(z.outBuffer))
		z.outBuffer = (*C.Byte)(C._null)
	}
}

func (z *SevenZip) Close() {
	z.flushBuffer()
	C.SzArEx_Free(&z.db, &z.allocImp)
	C.File_Close(&z.archiveStream.file)
}

//The whole *[0]uint8 thing is a mystery to me.

// Opens a 7-zip archive at a given path.
func Open(filename string) (*SevenZip, os.Error) {
	z := new(SevenZip)
	z.blockIndex = 0xffffffff
	z.allocImp.Alloc = (*[0]uint8)(C._szAlloc)
	z.allocImp.Free = (*[0]uint8)(C._szFree)
	z.allocTempImp.Alloc = (*[0]uint8)(C._szAllocTemp)
	z.allocTempImp.Free = (*[0]uint8)(C._szFreeTemp)

	cfilename := C.CString(filename)
	defer C.free(unsafe.Pointer(cfilename))
	if C.InFile_Open(&z.archiveStream.file, cfilename) != 0 {
		return nil, os.NewError("sevenzip: file not found: " + filename)
	}

	C.FileInStream_CreateVTable(&z.archiveStream)
	C.LookToRead_CreateVTable(&z.lookStream, 0)

	z.lookStream.realStream = &z.archiveStream.s
	C.LookToRead_Init(&z.lookStream)

	C.SzArEx_Init(&z.db)
	if C.SzArEx_Open(&z.db, &z.lookStream.s, &z.allocImp, &z.allocTempImp) != C._SZ_OK {
		return nil, os.NewError("sevenzip: failed to open archive: " + filename)
	}

	z.File = make([]*File, int(z.db.db.NumFiles))

	for i := 0; i < int(z.db.db.NumFiles); i++ {
		cf := C._CSzFileItemArrayIndex(z.db.db.Files, C.size_t(i))
		header := (*FileHeader)(unsafe.Pointer(cf))
		name := z.name(i)
		z.File[i] = &File{FileHeader: header, Name: name, Index: i, z: z}
	}

	return z, nil
}

// The byte array returned by this function is unsafe.
// It is mapped to a C buffer, which can be erased/overwritten
// when Close, ExtractUnsafe,or any function that calls ExtractUnsafe is called.
// Unless your data is too big, and you are prepared to
// avoid using another call to mentioned functions during the life-time of
// the returned array, do not use this function.
func (z *SevenZip) ExtractUnsafe(i int) ([]byte, os.Error) {
	var outSizeProcessed, offset C.size_t

	if C.SzArEx_Extract(
		&z.db, &z.lookStream.s, C.UInt32(i),
		&z.blockIndex, &z.outBuffer, &z.outBufferSize,
		&offset, &outSizeProcessed,
		&z.allocImp, &z.allocTempImp) != C._SZ_OK {
		return []byte{}, os.NewError("sevenzip: extract failed")
	}

	cbuf := C._ByteArrayIndex(z.outBuffer, offset)
	clen := int(outSizeProcessed)

	file := (*[1 << 30]byte)(unsafe.Pointer(cbuf))[:clen]
	hdr := (*reflect.SliceHeader)(unsafe.Pointer(&file))
	hdr.Cap = clen
	hdr.Len = clen

	return file, nil
}

// This function will call ExtractUnsafe, and copy the
// returned data to a new array which can be used
// safely as a normal array. Then the internal
// buffer is flushed.
// The catch is, this function, during it's execution,
// will use up double of the size of the ith file.
func (z *SevenZip) Extract(i int) ([]byte, os.Error) {
	z.l.Lock()
	defer z.l.Unlock()

	cfile, err := z.ExtractUnsafe(i)
	if err != nil {
		return []byte{}, err
	}
	file := make([]byte, len(cfile))
	copy(file, cfile)

	return file, nil
}
