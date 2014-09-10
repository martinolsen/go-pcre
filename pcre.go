package pcre

// #cgo LDFLAGS: -lpcre
// #include <pcre.h>
// #include <string.h>
//
// void
// call_pcre_free(void *ptr)
// {
//     pcre_free(ptr);
// }
import "C"

import (
	"errors"
	"unsafe"
	"reflect"
	"log"
)

const (
	Major int = C.PCRE_MAJOR
	Minor     = C.PCRE_MINOR
	Date      = C.PCRE_DATE
)

type Option int

const (
	Caseless         Option = C.PCRE_CASELESS
	Multiline               = C.PCRE_MULTILINE
	Dotall                  = C.PCRE_DOTALL
	Extended                = C.PCRE_EXTENDED
	Anchored                = C.PCRE_ANCHORED
	DollarEndonly           = C.PCRE_DOLLAR_ENDONLY
	Extra                   = C.PCRE_EXTRA
	Notbol                  = C.PCRE_NOTBOL
	Noteol                  = C.PCRE_NOTEOL
	Ungreedy                = C.PCRE_UNGREEDY
	Notempty                = C.PCRE_NOTEMPTY
	Utf8                    = C.PCRE_UTF8
	Utf16                   = C.PCRE_UTF16
	NoAutoCapture           = C.PCRE_NO_AUTO_CAPTURE
	NoUtf8Check             = C.PCRE_NO_UTF8_CHECK
	NoUtf16Check            = C.PCRE_NO_UTF16_CHECK
	AutoCallout             = C.PCRE_AUTO_CALLOUT
	PartialSoft             = C.PCRE_PARTIAL_SOFT
	Partial                 = C.PCRE_PARTIAL
	DfaShortest             = C.PCRE_DFA_SHORTEST
	DfaRestart              = C.PCRE_DFA_RESTART
	Firstline               = C.PCRE_FIRSTLINE
	Dupnames                = C.PCRE_DUPNAMES
	NewlineCr               = C.PCRE_NEWLINE_CR
	NewlineLf               = C.PCRE_NEWLINE_LF
	NewlineCrlf             = C.PCRE_NEWLINE_CRLF
	NewlineAny              = C.PCRE_NEWLINE_ANY
	NewlineAnycrlf          = C.PCRE_NEWLINE_ANYCRLF
	BsrAnycrlf              = C.PCRE_BSR_ANYCRLF
	BsrUnicode              = C.PCRE_BSR_UNICODE
	JavascriptCompat        = C.PCRE_JAVASCRIPT_COMPAT
	NoStartOptimize         = C.PCRE_NO_START_OPTIMIZE
	NoStartOptimise         = C.PCRE_NO_START_OPTIMISE
	PartialHard             = C.PCRE_PARTIAL_HARD
	NotemptyAtstart         = C.PCRE_NOTEMPTY_ATSTART
	Ucp                     = C.PCRE_UCP
)

type Info int

const (
	InfoOptions       Info = C.PCRE_INFO_OPTIONS
	InfoSize               = C.PCRE_INFO_SIZE
	InfoCapturecount       = C.PCRE_INFO_CAPTURECOUNT
	InfoBackrefmax         = C.PCRE_INFO_BACKREFMAX
	InfoFirstbyte          = C.PCRE_INFO_FIRSTBYTE
	InfoFirstchar          = C.PCRE_INFO_FIRSTCHAR // For backwards compatibility
	InfoFirsttable         = C.PCRE_INFO_FIRSTTABLE
	InfoLastliteral        = C.PCRE_INFO_LASTLITERAL
	InfoNameentrysize      = C.PCRE_INFO_NAMEENTRYSIZE
	InfoNamecount          = C.PCRE_INFO_NAMECOUNT
	InfoNametable          = C.PCRE_INFO_NAMETABLE
	InfoStudysize          = C.PCRE_INFO_STUDYSIZE
	InfoDefaultTables      = C.PCRE_INFO_DEFAULT_TABLES
	InfoOkpartial          = C.PCRE_INFO_OKPARTIAL
	InfoJchanged           = C.PCRE_INFO_JCHANGED
	InfoHascrorlf          = C.PCRE_INFO_HASCRORLF
	InfoMinlength          = C.PCRE_INFO_MINLENGTH
	InfoJit                = C.PCRE_INFO_JIT
	InfoJitsize            = C.PCRE_INFO_JITSIZE
	InfoMaxlookbehind      = C.PCRE_INFO_MAXLOOKBEHIND
)

type PCRE C.struct_real_pcre

func Compile(expr string, options Option, table interface{}) (*PCRE, error) {
	var (
		errptr    *C.char
		erroffset C.int
	)

	pattern := C.CString(expr)
	defer C.free(unsafe.Pointer(pattern))

	re := C.pcre_compile(pattern, C.int(options), &errptr, &erroffset, nil)
	if re == nil {
		return nil, errors.New(C.GoString(errptr))
	}

	return (*PCRE)(re), nil
}

func (pcre *PCRE) Free() { C.call_pcre_free(unsafe.Pointer(pcre)) }

func (pcre *PCRE) Exec(extra interface{}, subject string, startoffset int, options Option, ovector []int) Error {
	subjectCStr := C.CString(subject)
	defer C.free(unsafe.Pointer(subjectCStr))

	ovectorC := make([]C.int, len(ovector))
	for n, i := range ovector {
		ovectorC[n] = C.int(i)
	}

	var ovectorPtr *C.int
	if len(ovector) > 0 {
		ovectorPtr = &ovectorC[0]
	}

	r := C.pcre_exec((*C.struct_real_pcre)(pcre), nil, subjectCStr, C.int(len(subject)), C.int(startoffset), C.int(options), ovectorPtr, C.int(len(ovector)))

	for n, i := range ovectorC {
		ovector[n] = int(i)
	}

	return Error(r)
}

func (pcre *PCRE) Capturecount() int {
	var i C.int
	if rc := C.pcre_fullinfo((*C.struct_real_pcre)(pcre), nil, InfoCapturecount, unsafe.Pointer(&i)); rc != 0 {
		panic("pcre_fullinfo")
	}
	return int(i)
}

func (pcre *PCRE) Namecount() int {
	var i C.int
	if rc := C.pcre_fullinfo((*C.struct_real_pcre)(pcre), nil, InfoNamecount, unsafe.Pointer(&i)); rc != 0 {
		panic("pcre_fullinfo")
	}
	return int(i)
}

func (pcre *PCRE) Nameentrysize() int {
	var i C.int
	if rc := C.pcre_fullinfo((*C.struct_real_pcre)(pcre), nil, InfoNameentrysize, unsafe.Pointer(&i)); rc != 0 {
		panic("pcre_fullinfo")
	}
	return int(i)
}

func (pcre *PCRE) Nametable() []string {
	names := make([]string, pcre.Capturecount()+1)
	if pcre.Namecount() == 0 {
		return names
	}

	var dataPtr uintptr
	if rc := C.pcre_fullinfo((*C.struct_real_pcre)(pcre), nil, InfoNametable, unsafe.Pointer(&dataPtr)); rc != 0 {
		log.Panicf("pcre_fullinfo: %d", rc)
	}

	var data []byte = *(*[]byte)(unsafe.Pointer(
		&reflect.SliceHeader{
			Data: dataPtr,
			Len: pcre.Namecount() * pcre.Nameentrysize(),
			Cap: pcre.Namecount() * pcre.Nameentrysize(),
		}))

	for i := 0; i < len(data); {
		n := (int(data[i]) << 8) | int(data[i+1])
		s := string(data[i+2:i+pcre.Nameentrysize()-1])

		names[n] = s

		i += pcre.Nameentrysize()
	}

	return names
}

type Error int

const (
	ErrNomatch       Error = C.PCRE_ERROR_NOMATCH
	ErrNull                = C.PCRE_ERROR_NULL
	ErrBadoption           = C.PCRE_ERROR_BADOPTION
	ErrBadMagic            = C.PCRE_ERROR_BADMAGIC
	ErrUnknownOpcode       = C.PCRE_ERROR_UNKNOWN_OPCODE
)

/*
PCRE_ERROR_UNKNOWN_NODE
PCRE_ERROR_NOMEMORY
PCRE_ERROR_NOSUBSTRING
PCRE_ERROR_MATCHLIMIT
PCRE_ERROR_CALLOUT
PCRE_ERROR_BADUTF8
PCRE_ERROR_BADUTF16
PCRE_ERROR_BADUTF8_OFFSET
PCRE_ERROR_BADUTF16_OFFSET
PCRE_ERROR_PARTIAL
PCRE_ERROR_BADPARTIAL
PCRE_ERROR_INTERNAL
PCRE_ERROR_BADCOUNT
PCRE_ERROR_DFA_UITEM
PCRE_ERROR_DFA_UCOND
PCRE_ERROR_DFA_UMLIMIT
PCRE_ERROR_DFA_WSSIZE
PCRE_ERROR_DFA_RECURSE
PCRE_ERROR_RECURSIONLIMIT
PCRE_ERROR_NULLWSLIMIT
PCRE_ERROR_BADNEWLINE
PCRE_ERROR_BADOFFSET
PCRE_ERROR_SHORTUTF8
PCRE_ERROR_SHORTUTF16
PCRE_ERROR_RECURSELOOP
PCRE_ERROR_JIT_STACKLIMIT
PCRE_ERROR_BADMODE
PCRE_ERROR_BADENDIANNESS
PCRE_ERROR_DFA_BADRESTART

// Specific error codes for UTF-8 validity checks

#define PCRE_UTF8_ERR0               0
#define PCRE_UTF8_ERR1               1
#define PCRE_UTF8_ERR2               2
#define PCRE_UTF8_ERR3               3
#define PCRE_UTF8_ERR4               4
#define PCRE_UTF8_ERR5               5
#define PCRE_UTF8_ERR6               6
#define PCRE_UTF8_ERR7               7
#define PCRE_UTF8_ERR8               8
#define PCRE_UTF8_ERR9               9
#define PCRE_UTF8_ERR10             10
#define PCRE_UTF8_ERR11             11
#define PCRE_UTF8_ERR12             12
#define PCRE_UTF8_ERR13             13
#define PCRE_UTF8_ERR14             14
#define PCRE_UTF8_ERR15             15
#define PCRE_UTF8_ERR16             16
#define PCRE_UTF8_ERR17             17
#define PCRE_UTF8_ERR18             18
#define PCRE_UTF8_ERR19             19
#define PCRE_UTF8_ERR20             20
#define PCRE_UTF8_ERR21             21

// Specific error codes for UTF-16 validity checks

#define PCRE_UTF16_ERR0              0
#define PCRE_UTF16_ERR1              1
#define PCRE_UTF16_ERR2              2
#define PCRE_UTF16_ERR3              3
#define PCRE_UTF16_ERR4              4

// Request types for pcre_fullinfo()

// Request types for pcre_config(). Do not re-arrange, in order to remain compatible.

#define PCRE_CONFIG_UTF8                    0
#define PCRE_CONFIG_NEWLINE                 1
#define PCRE_CONFIG_LINK_SIZE               2
#define PCRE_CONFIG_POSIX_MALLOC_THRESHOLD  3
#define PCRE_CONFIG_MATCH_LIMIT             4
#define PCRE_CONFIG_STACKRECURSE            5
#define PCRE_CONFIG_UNICODE_PROPERTIES      6
#define PCRE_CONFIG_MATCH_LIMIT_RECURSION   7
#define PCRE_CONFIG_BSR                     8
#define PCRE_CONFIG_JIT                     9
#define PCRE_CONFIG_UTF16                  10
#define PCRE_CONFIG_JITTARGET              11

// Request types for pcre_study(). Do not re-arrange, in order to remain compatible.

#define PCRE_STUDY_JIT_COMPILE                0x0001
#define PCRE_STUDY_JIT_PARTIAL_SOFT_COMPILE   0x0002
#define PCRE_STUDY_JIT_PARTIAL_HARD_COMPILE   0x0004

// Bit flags for the pcre[16]_extra structure. Do not re-arrange or redefine these bits, just add new ones on the end, in order to remain compatible.

#define PCRE_EXTRA_STUDY_DATA             0x0001
#define PCRE_EXTRA_MATCH_LIMIT            0x0002
#define PCRE_EXTRA_CALLOUT_DATA           0x0004
#define PCRE_EXTRA_TABLES                 0x0008
#define PCRE_EXTRA_MATCH_LIMIT_RECURSION  0x0010
#define PCRE_EXTRA_MARK                   0x0020
#define PCRE_EXTRA_EXECUTABLE_JIT         0x0040
*/
