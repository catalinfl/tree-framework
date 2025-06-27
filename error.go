package tree

import "errors"

var (
	ErrInvalidJSONType          = errors.New("invalid json type")
	ErrInvalidJSONEncode        = errors.New("invalid json encode")
	ErrInvalidParam             = errors.New("invalid parameter")
	ErrParamInexistent          = errors.New("parameter does not exist")
	ErrQueryInexistent          = errors.New("query does not exist")
	ErrRegexParamDoesntExist    = errors.New("regex parameter does not exist")
	ErrRegexNotRespected        = errors.New("regex not respected")
	ErrKeysInexistent           = errors.New("keys map does not exist")
	ErrKeyNotFound              = errors.New("key not found")
	ErrInvalidType              = errors.New("invalid type")
	ErrInvalidCookieParam       = errors.New("invalid cookie parameter")
	ErrCookieNotFound           = errors.New("cookie not found")
	ErrHeaderNotFound           = errors.New("header not found")
	ErrOneOrMoreHeadersNotFound = errors.New("one or more headers not found")
)
