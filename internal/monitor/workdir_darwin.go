package monitor

// #include <libproc.h>
// #include <stdlib.h>
// #include <string.h>
import "C"

import (
	"fmt"
	"unsafe"
)

// getWorkingDir retrieves the current working directory of a process using macOS proc_pidinfo
func getWorkingDir(pid int32) (string, error) {
	// Allocate memory for proc_vnodepathinfo
	var pathInfo C.struct_proc_vnodepathinfo
	size := C.int(unsafe.Sizeof(pathInfo))

	// Call proc_pidinfo to get the working directory
	ret := C.proc_pidinfo(
		C.int(pid),
		C.PROC_PIDVNODEPATHINFO,
		0,
		unsafe.Pointer(&pathInfo),
		size,
	)

	if ret <= 0 {
		return "", fmt.Errorf("proc_pidinfo failed for PID %d: permission denied or process not found", pid)
	}

	// Extract the current working directory path from vip_cdir
	cwd := C.GoString(&pathInfo.pvi_cdir.vip_path[0])
	return cwd, nil
}
