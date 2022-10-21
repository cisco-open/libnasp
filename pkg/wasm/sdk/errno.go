// Copyright (c) 2022 Cisco and/or its affiliates. All rights reserved.
//
//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this file except in compliance with the License.
//  You may obtain a copy of the License at
//
//       https://www.apache.org/licenses/LICENSE-2.0
//
//  Unless required by applicable law or agreed to in writing, software
//  distributed under the License is distributed on an "AS IS" BASIS,
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//  See the License for the specific language governing permissions and
//  limitations under the License.

package sdk

import (
	"fmt"
)

// Errno are the error codes returned by WASI functions.
type Errno = uint32

// ErrnoName returns the POSIX error code name, except ErrnoSuccess, which is not an error. Ex. Errno2big -> "E2BIG"
func ErrnoName(errno Errno) string {
	if int(errno) < len(posixErrorCodes) {
		return posixErrorCodes[errno]
	}
	return fmt.Sprintf("errno(%d)", errno)
}

func ErrnoString(errno Errno) string {
	code := ErrnoName(errno)
	if s, ok := posixErrorStrings[code]; ok {
		return s
	}

	return code
}

// Posix errors
const (
	// ErrnoSuccess No error occurred. System call completed successfully.
	ErrnoSuccess Errno = iota
	// Errno2big Argument list too long.
	Errno2big
	// ErrnoAcces Permission denied.
	ErrnoAcces
	// ErrnoAddrinuse Address in use.
	ErrnoAddrinuse
	// ErrnoAddrnotavail Address not available.
	ErrnoAddrnotavail
	// ErrnoAfnosupport Address family not supported.
	ErrnoAfnosupport
	// ErrnoAgain Resource unavailable, or operation would block.
	ErrnoAgain
	// ErrnoAlready Connection already in progress.
	ErrnoAlready
	// ErrnoBadf Bad file descriptor.
	ErrnoBadf
	// ErrnoBadmsg Bad message.
	ErrnoBadmsg
	// ErrnoBusy Device or resource busy.
	ErrnoBusy
	// ErrnoCanceled Operation canceled.
	ErrnoCanceled
	// ErrnoChild No child processes.
	ErrnoChild
	// ErrnoConnaborted Connection aborted.
	ErrnoConnaborted
	// ErrnoConnrefused Connection refused.
	ErrnoConnrefused
	// ErrnoConnreset Connection reset.
	ErrnoConnreset
	// ErrnoDeadlk Resource deadlock would occur.
	ErrnoDeadlk
	// ErrnoDestaddrreq Destination address required.
	ErrnoDestaddrreq
	// ErrnoDom Mathematics argument out of domain of function.
	ErrnoDom
	// ErrnoDquot Reserved.
	ErrnoDquot
	// ErrnoExist File exists.
	ErrnoExist
	// ErrnoFault Bad address.
	ErrnoFault
	// ErrnoFbig File too large.
	ErrnoFbig
	// ErrnoHostunreach Host is unreachable.
	ErrnoHostunreach
	// ErrnoIdrm Identifier removed.
	ErrnoIdrm
	// ErrnoIlseq Illegal byte sequence.
	ErrnoIlseq
	// ErrnoInprogress Operation in progress.
	ErrnoInprogress
	// ErrnoIntr Interrupted function.
	ErrnoIntr
	// ErrnoInval Invalid argument.
	ErrnoInval
	// ErrnoIo I/O error.
	ErrnoIo
	// ErrnoIsconn Socket is connected.
	ErrnoIsconn
	// ErrnoIsdir Is a directory.
	ErrnoIsdir
	// ErrnoLoop Too many levels of symbolic links.
	ErrnoLoop
	// ErrnoMfile File descriptor value too large.
	ErrnoMfile
	// ErrnoMlink Too many links.
	ErrnoMlink
	// ErrnoMsgsize Message too large.
	ErrnoMsgsize
	// ErrnoMultihop Reserved.
	ErrnoMultihop
	// ErrnoNametoolong Filename too long.
	ErrnoNametoolong
	// ErrnoNetdown Network is down.
	ErrnoNetdown
	// ErrnoNetreset Connection aborted by network.
	ErrnoNetreset
	// ErrnoNetunreach Network unreachable.
	ErrnoNetunreach
	// ErrnoNfile Too many files open in system.
	ErrnoNfile
	// ErrnoNobufs No buffer space available.
	ErrnoNobufs
	// ErrnoNodev No such device.
	ErrnoNodev
	// ErrnoNoent No such file or directory.
	ErrnoNoent
	// ErrnoNoexec Executable file format error.
	ErrnoNoexec
	// ErrnoNolck No locks available.
	ErrnoNolck
	// ErrnoNolink Reserved.
	ErrnoNolink
	// ErrnoNomem Not enough space.
	ErrnoNomem
	// ErrnoNomsg No message of the desired type.
	ErrnoNomsg
	// ErrnoNoprotoopt No message of the desired type.
	ErrnoNoprotoopt
	// ErrnoNospc No space left on device.
	ErrnoNospc
	// ErrnoNosys function not supported.
	ErrnoNosys
	// ErrnoNotconn The socket is not connected.
	ErrnoNotconn
	// ErrnoNotdir Not a directory or a symbolic link to a directory.
	ErrnoNotdir
	// ErrnoNotempty Directory not empty.
	ErrnoNotempty
	// ErrnoNotrecoverable State not recoverable.
	ErrnoNotrecoverable
	// ErrnoNotsock Not a socket.
	ErrnoNotsock
	// ErrnoNotsup Not supported, or operation not supported on socket.
	ErrnoNotsup
	// ErrnoNotty Inappropriate I/O control operation.
	ErrnoNotty
	// ErrnoNxio No such device or address.
	ErrnoNxio
	// ErrnoOverflow Value too large to be stored in data type.
	ErrnoOverflow
	// ErrnoOwnerdead Previous owner died.
	ErrnoOwnerdead
	// ErrnoPerm Operation not permitted.
	ErrnoPerm
	// ErrnoPipe Broken pipe.
	ErrnoPipe
	// ErrnoProto Protocol error.
	ErrnoProto
	// ErrnoProtonosupport Protocol error.
	ErrnoProtonosupport
	// ErrnoPrototype Protocol wrong type for socket.
	ErrnoPrototype
	// ErrnoRange Result too large.
	ErrnoRange
	// ErrnoRofs Read-only file system.
	ErrnoRofs
	// ErrnoSpipe Invalid seek.
	ErrnoSpipe
	// ErrnoSrch No such process.
	ErrnoSrch
	// ErrnoStale Reserved.
	ErrnoStale
	// ErrnoTimedout Connection timed out.
	ErrnoTimedout
	// ErrnoTxtbsy Text file busy.
	ErrnoTxtbsy
	// ErrnoXdev Cross-device link.
	ErrnoXdev
	// ErrnoNotcapable Extension: Capabilities insufficient.
	ErrnoNotcapable
)

var posixErrorCodes = [...]string{
	ErrnoSuccess:        "ESUCCESS",
	Errno2big:           "E2BIG",
	ErrnoAcces:          "EACCES",
	ErrnoAddrinuse:      "EADDRINUSE",
	ErrnoAddrnotavail:   "EADDRNOTAVAIL",
	ErrnoAfnosupport:    "EAFNOSUPPORT",
	ErrnoAgain:          "EAGAIN",
	ErrnoAlready:        "EALREADY",
	ErrnoBadf:           "EBADF",
	ErrnoBadmsg:         "EBADMSG",
	ErrnoBusy:           "EBUSY",
	ErrnoCanceled:       "ECANCELED",
	ErrnoChild:          "ECHILD",
	ErrnoConnaborted:    "ECONNABORTED",
	ErrnoConnrefused:    "ECONNREFUSED",
	ErrnoConnreset:      "ECONNRESET",
	ErrnoDeadlk:         "EDEADLK",
	ErrnoDestaddrreq:    "EDESTADDRREQ",
	ErrnoDom:            "EDOM",
	ErrnoDquot:          "EDQUOT",
	ErrnoExist:          "EEXIST",
	ErrnoFault:          "EFAULT",
	ErrnoFbig:           "EFBIG",
	ErrnoHostunreach:    "EHOSTUNREACH",
	ErrnoIdrm:           "EIDRM",
	ErrnoIlseq:          "EILSEQ",
	ErrnoInprogress:     "EINPROGRESS",
	ErrnoIntr:           "EINTR",
	ErrnoInval:          "EINVAL",
	ErrnoIo:             "EIO",
	ErrnoIsconn:         "EISCONN",
	ErrnoIsdir:          "EISDIR",
	ErrnoLoop:           "ELOOP",
	ErrnoMfile:          "EMFILE",
	ErrnoMlink:          "EMLINK",
	ErrnoMsgsize:        "EMSGSIZE",
	ErrnoMultihop:       "EMULTIHOP",
	ErrnoNametoolong:    "ENAMETOOLONG",
	ErrnoNetdown:        "ENETDOWN",
	ErrnoNetreset:       "ENETRESET",
	ErrnoNetunreach:     "ENETUNREACH",
	ErrnoNfile:          "ENFILE",
	ErrnoNobufs:         "ENOBUFS",
	ErrnoNodev:          "ENODEV",
	ErrnoNoent:          "ENOENT",
	ErrnoNoexec:         "ENOEXEC",
	ErrnoNolck:          "ENOLCK",
	ErrnoNolink:         "ENOLINK",
	ErrnoNomem:          "ENOMEM",
	ErrnoNomsg:          "ENOMSG",
	ErrnoNoprotoopt:     "ENOPROTOOPT",
	ErrnoNospc:          "ENOSPC",
	ErrnoNosys:          "ENOSYS",
	ErrnoNotconn:        "ENOTCONN",
	ErrnoNotdir:         "ENOTDIR",
	ErrnoNotempty:       "ENOTEMPTY",
	ErrnoNotrecoverable: "ENOTRECOVERABLE",
	ErrnoNotsock:        "ENOTSOCK",
	ErrnoNotsup:         "ENOTSUP",
	ErrnoNotty:          "ENOTTY",
	ErrnoNxio:           "ENXIO",
	ErrnoOverflow:       "EOVERFLOW",
	ErrnoOwnerdead:      "EOWNERDEAD",
	ErrnoPerm:           "EPERM",
	ErrnoPipe:           "EPIPE",
	ErrnoProto:          "EPROTO",
	ErrnoProtonosupport: "EPROTONOSUPPORT",
	ErrnoPrototype:      "EPROTOTYPE",
	ErrnoRange:          "ERANGE",
	ErrnoRofs:           "EROFS",
	ErrnoSpipe:          "ESPIPE",
	ErrnoSrch:           "ESRCH",
	ErrnoStale:          "ESTALE",
	ErrnoTimedout:       "ETIMEDOUT",
	ErrnoTxtbsy:         "ETXTBSY",
	ErrnoXdev:           "EXDEV",
	ErrnoNotcapable:     "ENOTCAPABLE",
}

var posixErrorStrings = map[string]string{
	"E2BIG":           "argument list too long",
	"EACCES":          "permission denied",
	"EADDRINUSE":      "address already in use",
	"EADDRNOTAVAIL":   "address not available",
	"EAFNOSUPPORT":    "address family not supported",
	"EAGAIN":          "resource temporarily unavailable",
	"EALREADY":        "connection already in progress",
	"EBADE":           "invalid exchange",
	"EBADF":           "bad file descriptor",
	"EBADFD":          "file descriptor in bad state",
	"EBADMSG":         "bad message",
	"EBADR":           "invalid request descriptor",
	"EBADRQC":         "invalid request code",
	"EBADSLT":         "invalid slot",
	"EBUSY":           "device or resource busy",
	"ECANCELED":       "operation canceled",
	"ECHILD":          "no child processes",
	"ECHRNG":          "channel number out of range",
	"ECOMM":           "communication error on send",
	"ECONNABORTED":    "connection aborted",
	"ECONNREFUSED":    "connection refused",
	"ECONNRESET":      "connection reset",
	"EDEADLK":         "resource deadlock avoided",
	"EDESTADDRREQ":    "destination address required",
	"EDOM":            "mathematics argument out of domain of function",
	"EDQUOT":          "disk quota exceeded",
	"EEXIST":          "file exists",
	"EFAULT":          "bad address",
	"EFBIG":           "file too large",
	"EHOSTDOWN":       "host is down",
	"EHOSTUNREACH":    "host is unreachable",
	"EIDRM":           "identifier removed",
	"EILSEQ":          "illegal byte sequence (posix.1, C99)",
	"EINPROGRESS":     "operation in progress",
	"EINTR":           "interrupted function call; see signal(7).",
	"EINVAL":          "invalid argument",
	"EIO":             "input/output error",
	"EISCONN":         "socket is connected",
	"EISDIR":          "is a directory",
	"EISNAM":          "is a named type file",
	"EKEYEXPIRED":     "key has expired",
	"EKEYREJECTED":    "key was rejected by service",
	"EKEYREVOKED":     "key has been revoked",
	"EL2HLT":          "level 2 halted",
	"EL2NSYNC":        "level 2 not synchronized",
	"EL3HLT":          "level 3 halted",
	"EL3RST":          "level 3 halted",
	"ELIBACC":         "cannot access a needed shared library",
	"ELIBBAD":         "accessing a corrupted shared library",
	"ELIBMAX":         "attempting to link in too many shared libraries",
	"ELIBSCN":         "lib section in a.out corrupted",
	"ELIBEXEC":        "cannot exec a shared library directly",
	"ELOOP":           "too many levels of symbolic links",
	"EMEDIUMTYPE":     "wrong medium type",
	"EMFILE":          "too many open files",
	"EMLINK":          "too many links",
	"EMSGSIZE":        "message too long",
	"EMULTIHOP":       "multihop attempted",
	"ENAMETOOLONG":    "filename too long",
	"ENETDOWN":        "network is down",
	"ENETRESET":       "connection aborted by network",
	"ENETUNREACH":     "network unreachable",
	"ENFILE":          "too many open files in system",
	"ENOBUFS":         "no buffer space available",
	"ENODATA":         "no message is available on the sTREAM head read queue",
	"ENODEV":          "no such device",
	"ENOENT":          "no such file or directory",
	"ENOEXEC":         "exec format error",
	"ENOKEY":          "required key not available",
	"ENOLCK":          "no locks available",
	"ENOLINK":         "link has been severed",
	"ENOMEDIUM":       "no medium found",
	"ENOMEM":          "not enough space",
	"ENOMSG":          "no message of the desired type",
	"ENONET":          "machine is not on the network",
	"ENOPKG":          "package not installed",
	"ENOPROTOOPT":     "protocol not available",
	"ENOSPC":          "no space left on device",
	"ENOSR":           "no stream resources",
	"ENOSTR":          "not a stream",
	"ENOSYS":          "function not implemented",
	"ENOTBLK":         "block device required",
	"ENOTCONN":        "the socket is not connected",
	"ENOTDIR":         "not a directory",
	"ENOTEMPTY":       "directory not empty",
	"ENOTSOCK":        "not a socket",
	"ENOTSUP":         "operation not supported",
	"ENOTTY":          "inappropriate i/o control operation",
	"ENOTUNIQ":        "name not unique on network",
	"ENXIO":           "no such device or address",
	"EOPNOTSUPP":      "operation not supported on socket",
	"EOVERFLOW":       "value too large to be stored in data type",
	"EPERM":           "operation not permitted",
	"EPFNOSUPPORT":    "protocol family not supported",
	"EPIPE":           "broken pipe",
	"EPROTO":          "protocol error",
	"EPROTONOSUPPORT": "protocol not supported",
	"EPROTOTYPE":      "protocol wrong type for socket",
	"ERANGE":          "result too large",
	"EREMCHG":         "remote address changed",
	"EREMOTE":         "object is remote",
	"EREMOTEIO":       "remote i/o error",
	"ERESTART":        "interrupted system call should be restarted",
	"EROFS":           "read-only file system",
	"ESHUTDOWN":       "cannot send after transport endpoint shutdown",
	"ESPIPE":          "invalid seek",
	"ESOCKTNOSUPPORT": "socket type not supported",
	"ESRCH":           "no such process",
	"ESTALE":          "stale file handle",
	"ESTRPIPE":        "streams pipe error",
	"ETIME":           "timer expired",
	"ETIMEDOUT":       "connection timed out",
	"ETXTBSY":         "text file busy",
	"EUCLEAN":         "structure needs cleaning",
	"EUNATCH":         "protocol driver not attached",
	"EUSERS":          "too many users",
	"EWOULDBLOCK":     "operation would block",
	"EXDEV":           "improper link",
	"EXFULL":          "exchange full",
}
