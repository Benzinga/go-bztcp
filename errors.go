package bztcp

import "fmt"

type (
	// InvalidReadyRespError occurs when the READY message does not occur as
	// expected (BZTCP v1+ protocol)
	InvalidReadyRespError struct{}

	// InvalidAuthResponseError occurs when the response to the AUTH message is
	// an unknown status
	InvalidAuthResponseError struct{}

	// InvalidKeyFormatError occurs when the server can't parse our key,
	// usually as a result of a programming error.
	InvalidKeyFormatError struct{}

	// InvalidKeyError occurs when the key passed to the TCP server is
	// not valid.
	InvalidKeyError struct{}

	// UnexpectedByteError occurs when the STREAM message contains an
	// unexpected byte in the Tickers field (BZTCP v1.1+ protocol)
	UnexpectedByteError byte
)

var (
	// ErrInvalidReady is a static instance of InvalidReadyRespError.
	ErrInvalidReady = InvalidReadyRespError{}

	// ErrInvalidAuthResponse is a static instance of InvalidAuthResponseError.
	ErrInvalidAuthResponse = InvalidAuthResponseError{}

	// ErrInvalidKeyFormat is a static instance of InvalidKeyFormatError.
	ErrInvalidKeyFormat = InvalidKeyFormatError{}

	// ErrInvalidKey is a static instance of InvalidKeyError.
	ErrInvalidKey = InvalidKeyError{}
)

// Error implements the error interface.
func (InvalidReadyRespError) Error() string {
	return "invalid ready message"
}

// Error implements the error interface.
func (InvalidAuthResponseError) Error() string {
	return "invalid auth response"
}

// Error implements the error interface.
func (InvalidKeyFormatError) Error() string {
	return "invalid key format"
}

// Error implements the error interface.
func (InvalidKeyError) Error() string {
	return "invalid key"
}

// Error implements the error interface.
func (b UnexpectedByteError) Error() string {
	return fmt.Sprintf("unexpected byte '%c'", byte(b))
}
