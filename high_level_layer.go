/*
Unless explicitly stated otherwise all files in this repository are licensed
under the MIT License.
This product includes software developed at Datadog (https://www.datadoghq.com/).
Copyright 2018 Datadog, Inc.
*/

package python3

/*
#cgo pkg-config: python3
#include "Python.h"
*/
import "C"
import (
	"fmt"
	"unsafe"
)

//Py_Main : https://docs.python.org/3/c-api/veryhigh.html?highlight=pycompilerflags#c.Py_Main
// "error" will be set if we fail to call "Py_DecodeLocale" on every "args".
func Py_Main(args []string) (int, error) {
	argc := C.int(len(args))
	argv := make([]*C.wchar_t, argc, argc)
	for i, arg := range args {
		carg := C.CString(arg)
		defer C.free(unsafe.Pointer(carg))

		warg := C.Py_DecodeLocale(carg, nil)
		if warg == nil {
			return -1, fmt.Errorf("fail to call Py_DecodeLocale on '%s'", arg)
		}
		// Py_DecodeLocale requires a call to PyMem_RawFree to free the memory
		defer C.PyMem_RawFree(unsafe.Pointer(warg))
		argv[i] = warg
	}

	return int(C.Py_Main(argc, (**C.wchar_t)(unsafe.Pointer(&argv[0])))), nil
}

//PyRun_AnyFile : https://docs.python.org/3/c-api/veryhigh.html?highlight=pycompilerflags#c.PyRun_AnyFile
// "error" will be set if we fail to open "filename".
func PyRun_AnyFile(filename string) (int, error) {
	cfilename := C.CString(filename)
	defer C.free(unsafe.Pointer(cfilename))

	mode := C.CString("r")
	defer C.free(unsafe.Pointer(mode))

	cfile, err := C.fopen(cfilename, mode)
	if err != nil {
		return -1, fmt.Errorf("fail to open '%s': %s", filename, err)
	}
	defer C.fclose(cfile)

	// C.PyRun_AnyFile is a macro, using C.PyRun_AnyFileFlags instead
	return int(C.PyRun_AnyFileFlags(cfile, cfilename, nil)), nil
}

//PyRun_SimpleString : https://docs.python.org/3/c-api/veryhigh.html?highlight=pycompilerflags#c.PyRun_SimpleString
func PyRun_SimpleString(command string) int {
	ccommand := C.CString(command)
	defer C.free(unsafe.Pointer(ccommand))

	// C.PyRun_SimpleString is a macro, using C.PyRun_SimpleStringFlags instead
	return int(C.PyRun_SimpleStringFlags(ccommand, nil))
}

//PyRun_String : https://docs.python.org/3/c-api/veryhigh.html?highlight=pycompilerflags#c.PyRun_String
func PyRun_String(command string, locals map[string]interface{}) *PyObject {
	ccommand := C.CString(command)
	defer C.free(unsafe.Pointer(ccommand))
	m := PyImport_AddModule("__main__")
	defer m.DecRef()
	globalDict := PyModule_GetDict(m)
	defer globalDict.DecRef()
	localDict := PyModule_GetDict(m)
	defer localDict.DecRef()

	for k, v := range locals {
		var item *PyObject
		switch parsed := v.(type) {
		case int:
			item = PyLong_FromGoInt(parsed)
		case bool:
			if parsed {
				item = Py_True
				break
			}
			item = Py_False
		case string:
			item = PyUnicode_FromString(parsed)
		}
		if item != nil {
			PyDict_SetItemString(localDict, k, item)
			defer item.DecRef()
		}
	}

	return togo(C.PyRun_StringFlags(ccommand, C.Py_file_input, toc(globalDict), toc(localDict), nil))
}
