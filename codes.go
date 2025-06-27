package tree

const (
	// Informational
	Continue           = 100
	SwitchingProtocols = 101
	Processing         = 102
	EarlyHints         = 103

	// Success
	OK                   = 200
	Created              = 201
	Accepted             = 202
	NonAuthoritativeInfo = 203
	NoContent            = 204
	ResetContent         = 205
	PartialContent       = 206
	MultiStatus          = 207
	AlreadyReported      = 208
	IMUsed               = 226

	// Redirection
	Multiple     = 300
	Moved        = 301
	Found        = 302
	SeeOther     = 303
	NotModified  = 304
	UseProxy     = 305
	TempRedirect = 307
	PermRedirect = 308

	// Client Error
	BadRequest                   = 400
	Unauthorized                 = 401
	PaymentRequired              = 402
	Forbidden                    = 403
	NotFound                     = 404
	MethodNotAllowed             = 405
	NotAcceptable                = 406
	ProxyAuthRequired            = 407
	RequestTimeout               = 408
	Conflict                     = 409
	Gone                         = 410
	LengthRequired               = 411
	PreconditionFailed           = 412
	RequestEntityTooLarge        = 413
	RequestURITooLong            = 414
	UnsupportedMediaType         = 415
	RequestedRangeNotSatisfiable = 416
	ExpectationFailed            = 417
	Teapot                       = 418
	MisdirectedRequest           = 421
	UnprocessableEntity          = 422
	Locked                       = 423
	FailedDependency             = 424
	TooEarly                     = 425
	UpgradeRequired              = 426
	PreconditionRequired         = 428
	ManyRequests                 = 429
	RequestHeaderFieldsTooLarge  = 431
	UnavailableForLegalReasons   = 451

	// Server Error
	InternalError                 = 500
	NotImplemented                = 501
	BadGateway                    = 502
	ServiceUnavailable            = 503
	GatewayTimeout                = 504
	HTTPVersionNotSuported        = 505
	VariantAlsoNegotiates         = 506
	InsufficientStorage           = 507
	LoopDetected                  = 508
	NotExtended                   = 510
	NetworkAuthenticationRequired = 511
)
