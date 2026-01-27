//go:build windows
// +build windows

// pkg/utils/winver_windows.go
// Minimal Win32 wrapper to read a file's version from its VS_FIXEDFILEINFO.
// Returns "a.b.c.d" or an error. Used to enrich note.json at build time.

package utils

import (
	"fmt"
	"syscall"
	"unsafe"
)

var (
	versionDll                 = syscall.NewLazyDLL("version.dll")
	procGetFileVersionInfoSize = versionDll.NewProc("GetFileVersionInfoSizeW")
	procGetFileVersionInfo     = versionDll.NewProc("GetFileVersionInfoW")
	procVerQueryValue          = versionDll.NewProc("VerQueryValueW")
)

// VS_FIXEDFILEINFO mirrors Win32 struct layout.
type VS_FIXEDFILEINFO struct {
	Signature        uint32
	StrucVersion     uint32
	FileVersionMS    uint32
	FileVersionLS    uint32
	ProductVersionMS uint32
	ProductVersionLS uint32
	FileFlagsMask    uint32
	FileFlags        uint32
	FileOS           uint32
	FileType         uint32
	FileSubtype      uint32
	FileDateMS       uint32
	FileDateLS       uint32
}

// ReadFileVersion returns the FileVersion (a.b.c.d) of a PE file.
func ReadFileVersion(path string) (string, error) {
	p16, err := syscall.UTF16PtrFromString(path)
	if err != nil {
		return "", err
	}

	// size query
	var zero uintptr
	r0, _, _ := procGetFileVersionInfoSize.Call(uintptr(unsafe.Pointer(p16)), uintptr(unsafe.Pointer(&zero)))
	if r0 == 0 {
		return "", fmt.Errorf("GetFileVersionInfoSizeW failed for %s", path)
	}
	bufSize := uint32(r0)

	// allocate buffer and get version info
	buf := make([]byte, bufSize)
	ok, _, _ := procGetFileVersionInfo.Call(
		uintptr(unsafe.Pointer(p16)),
		0,
		uintptr(bufSize),
		uintptr(unsafe.Pointer(&buf[0])),
	)
	if ok == 0 {
		return "", fmt.Errorf("GetFileVersionInfoW failed for %s", path)
	}

	// query root block
	var blockPtr uintptr
	var blockLen uint32
	subBlock, _ := syscall.UTF16PtrFromString(`\`)
	ok, _, _ = procVerQueryValue.Call(
		uintptr(unsafe.Pointer(&buf[0])),
		uintptr(unsafe.Pointer(subBlock)),
		uintptr(unsafe.Pointer(&blockPtr)),
		uintptr(unsafe.Pointer(&blockLen)),
	)
	if ok == 0 || blockPtr == 0 {
		return "", fmt.Errorf("VerQueryValueW failed for %s", path)
	}

	info := (*VS_FIXEDFILEINFO)(unsafe.Pointer(blockPtr))
	// Compose a.b.c.d from hi/lo words.
	a := hiWord(info.FileVersionMS)
	b := loWord(info.FileVersionMS)
	c := hiWord(info.FileVersionLS)
	d := loWord(info.FileVersionLS)
	return fmt.Sprintf("%d.%d.%d.%d", a, b, c, d), nil
}

func hiWord(v uint32) uint16 { return uint16(v >> 16) }
func loWord(v uint32) uint16 { return uint16(v & 0xFFFF) }
