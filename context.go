package tree

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"net/url"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/catalinfl/tree-framework/binding"
	"github.com/catalinfl/tree-framework/render"
)

type Ctx struct {
	r               *http.Request
	w               http.ResponseWriter
	routerPath      string
	keys            map[string]any
	params          map[string]string
	middlewareIndex int
	middlewares     []Middleware
	handler         CtxFunc
	automatic       bool
	maxMemory       int64
	formParsed      bool
}

func NewCtx(
	w http.ResponseWriter,
	r *http.Request,
	urlServer string,
	middlewares []Middleware,
	handler CtxFunc,
	automatic bool) *Ctx {

	urlPathClient := r.URL.Path

	params := getParams(urlPathClient, urlServer)

	return &Ctx{
		r:               r,
		w:               w,
		routerPath:      urlServer,
		keys:            make(map[string]any),
		params:          params,
		middlewareIndex: -1,
		middlewares:     middlewares,
		handler:         handler,
		automatic:       automatic,
		maxMemory:       10 << 20, // 10MB
		formParsed:      false,
	}
}

type SameSite int

const (
	DEFAULT SameSite = iota
	NONE
	LAX
	STRICT
)

type acceptSpecial struct {
	acceptHeaderValue string
	q                 float64
}

type acceptHeader struct {
	acceptHeaderValue string
	variants          []string
	q                 float64
}

type ProtoInfo struct {
	Proto      string
	ProtoMajor int
	ProtoMinor int
}

type Cookie struct {
	Name       string
	Value      string
	Path       string
	Domain     string
	RawExpires string
	MaxAge     int
	Secure     bool
	HttpOnly   bool
	SameSite   SameSite
}

type TreeFile struct {
	MultipartFile       multipart.File
	MultipartFileHeader *multipart.FileHeader
}

type CtxFunc func(c *Ctx) error

func (c *Ctx) SendJSON(jsonStruct J, code int) error {
	c.SetHeader("Content-Type", "application/json")
	c.w.WriteHeader(code)
	return jsonStruct.encodeJSON(c.w)
}

func (c *Ctx) SendString(message string, code int) error {
	c.w.WriteHeader(code)
	_, err := c.w.Write([]byte(message))
	if err != nil {
		return fmt.Errorf("failed to write response: %w", err)
	}

	return nil
}

func (c *Ctx) SendFile(filePath string) error {
	http.ServeFile(c.w, c.r, filePath)
	return nil
}

func (c *Ctx) Redirect(url string, code int) error {
	if code < 300 || code > 308 {
		code = Found
	}

	http.Redirect(c.w, c.r, url, code)
	return nil
}

func (c *Ctx) Status(code int) error {
	c.w.WriteHeader(code)
	return nil
}

func (c *Ctx) Next() error {
	c.middlewareIndex++
	if c.middlewareIndex < len(c.middlewares) {
		middleware := c.middlewares[c.middlewareIndex]
		if middleware.Handler != nil {
			err := middleware.Handler(c)
			if err != nil {
				return fmt.Errorf("middleware error: %w", err)
			}
		}
	}

	return nil
}

// GetNext - Get the current middleware index
//
// Note: This function works only if automatic middleware is disabled. It returns -1 if automatic middleware is true
func (c *Ctx) GetNext() int {
	if c.automatic {
		fmt.Printf("[ERROR] Automatic middleware is enabled, this function is not available \n")
		return -1
	}
	if c.middlewareIndex < len(c.middlewares) {
		return c.middlewareIndex
	}

	return -1
}

// GetTotalMiddlewares - Get the total number of middlewares available
//
// Note: This function works only if automatic middleware is disabled
func (c *Ctx) GetTotalMiddlewares() int {
	return len(c.middlewares)
}

// HeaderSent returns the header map to be sent by the server
// to the client response.
//
// Note: http.Header is type of map[string][]string
func (c *Ctx) HeaderSent() http.Header {
	return c.w.Header()
}

// Header returns the header map received *by* the server
// from the client request.
//
// Note: http.Header is type of map[string][]string
func (c *Ctx) Header() http.Header {
	return c.r.Header
}

func (c *Ctx) Scheme() string {
	if c.r.TLS != nil {
		return "https"
	}

	return "http"
}

func (c *Ctx) GetURLParam(param string) (string, error) {
	urlPath := c.r.URL.Path
	routerPath := c.routerPath

	urlSplit := strings.Split(urlPath, "/")
	routerSplit := strings.Split(routerPath, "/")

	userParam := fmt.Sprintf(":%s", param)

	var idParamIndex = -1

	for i, rpath := range routerSplit {
		if rpath == userParam && strings.Contains(routerPath, userParam) {
			idParamIndex = i
			break
		}
	}

	if idParamIndex != -1 {
		valueToGet := urlSplit[idParamIndex]
		return valueToGet, nil
	} else {
		return "", ErrInvalidParam
	}
}

func (c *Ctx) GetParamInt(param string) (int, error) {
	paramValue, err := c.GetURLParam(param)
	if err != nil {
		return 0, err
	}

	intValue, err := strconv.Atoi(paramValue)
	if err != nil {
		return 0, fmt.Errorf("invalid parameter value for %s: %w", param, err)
	}
	return intValue, nil
}

func (c *Ctx) GetAllParamsInt() (map[string]int, error) {
	stringParams, err := c.GetAllParams()
	if err != nil {
		return nil, err
	}

	intParams := make(map[string]int, len(stringParams))
	var conversionErrors []string

	for key, value := range stringParams {
		intValue, err := strconv.Atoi(value)
		if err != nil {
			conversionErrors = append(conversionErrors, fmt.Sprintf("invalid parameter value for %s: %s", key, err.Error()))
			continue
		}

		intParams[key] = intValue
	}

	if len(conversionErrors) > 0 {
		return nil, fmt.Errorf("conversion errors: %s", strings.Join(conversionErrors, ", "))
	}

	if len(intParams) == 0 {
		return nil, errors.New("no valid integer parameters found")
	}

	return intParams, nil
}

func (c *Ctx) RegexURLParam(regexIndexToBeSearched int) (string, error) {
	urlPath := c.r.URL.Path
	routerPath := c.routerPath

	if regexIndexToBeSearched <= 0 {
		return "", errors.New("regex index must be positive")
	}

	routerSegments := strings.Split(routerPath, "/")
	urlSegments := strings.Split(urlPath, "/")

	regexCount := 0
	regexFound := false
	regexIndex := -1
	regexPattern := ""

	for i, segment := range routerSegments {
		if len(segment) == 0 {
			continue
		}

		if strings.HasPrefix(segment, ":|") && strings.HasSuffix(segment, "|") {
			regexCount++

			if regexCount == regexIndexToBeSearched {
				regexPattern = segment[2 : len(segment)-1]
				regexIndex = i
				regexFound = true
				break
			}
		}
	}

	if !regexFound || regexPattern == "" {
		return "", ErrRegexParamDoesntExist
	}

	if regexIndex < 0 || regexIndex >= len(urlSegments) {
		return "", ErrRegexParamDoesntExist
	}

	paramValue := urlSegments[regexIndex]

	if !useRegex(regexPattern, paramValue) {
		return "", ErrRegexNotRespected
	}

	return paramValue, nil
}

/* Ignore regex params */
func (c *Ctx) GetAllParams() (map[string]string, error) {
	params := make(map[string]string)
	for key, value := range c.params {
		if !strings.HasPrefix(key, "|") && (!strings.HasSuffix(key, "|") || !hasNumberSuffix(key)) {
			params[key] = value
		}
	}

	if len(params) == 0 {
		return nil, errors.New("no params found")
	}

	return params, nil
}

func (c *Ctx) GetAllRegexParams() (map[string]string, error) {
	params := make(map[string]string)

	for key, value := range c.params {
		if strings.HasPrefix(key, "|") && (strings.HasSuffix(key, "|") || hasNumberSuffix(key)) {
			verifyRegexWord := useRegex(key[1:len(key)-1], value)
			if !verifyRegexWord {
				return nil, errors.New("regex param is not correct")
			}
			params[key] = value
		}
	}

	if len(params) == 0 {
		return nil, errors.New("no regex params found")
	}

	return params, nil
}

func hasNumberSuffix(input string) bool {
	pattern := `\|_\d+$`
	matched, _ := regexp.MatchString(pattern, input)
	return matched
}

func useRegex(regex string, word string) bool {
	re := regexp.MustCompile(regex)
	match := re.FindString(word)
	return match == word
}

func (c *Ctx) GetQuery(query string) (string, error) {
	queryValue := c.r.URL.Query().Get(query)

	// http://localhost:8080/test?query=1&test=2

	if queryValue == "" {
		return "", ErrQueryInexistent
	}

	return queryValue, nil
}

func (c *Ctx) GetQueryInt(query string) (int, error) {
	queryValue, err := c.GetQuery(query)
	if err != nil {
		return 0, err
	}

	intValue, err := strconv.Atoi(queryValue)
	if err != nil {
		return 0, fmt.Errorf("invalid query value for %s: %w", query, err)
	}

	return intValue, nil
}

func (c *Ctx) GetAllQueriesMap(queries ...string) (map[string]string, error) {
	queryValues := make(map[string]string)
	var queryErrorArg []string
	for _, query := range queries {
		queryValue := c.r.URL.Query().Get(query)
		if queryValue == "" {
			queryErrorArg = append(queryErrorArg, query)
		} else {
			queryValues[query] = queryValue
		}
	}

	if len(queryErrorArg) > 0 {
		var s string = "next queries are inexistent:"

		for i, query := range queryErrorArg {
			if i == len(queryErrorArg)-1 {
				s += fmt.Sprintf(" %s.", query)
			} else {
				s += fmt.Sprintf(" %s", query)
			}
		}
	}

	if len(queryValues) == 0 {
		return nil, errors.New("no query found")
	}

	if len(queryValues) != len(queries) {
		return nil, fmt.Errorf("one or more queries are not found: %s", strings.Join(queryErrorArg, ", "))
	}
	return queryValues, nil
}

func (c *Ctx) GetAllQueries(queries ...string) ([]string, error) {
	var queryValues []string
	var queryErrorArg []string

	for _, query := range queries {
		queryValue := c.r.URL.Query().Get(query)

		if queryValue == "" {
			queryErrorArg = append(queryErrorArg, query)
		} else {
			queryValues = append(queryValues, queryValue)
		}
	}

	if len(queryErrorArg) > 0 {
		var s string = "next queries are inexistent:"

		for i, query := range queryErrorArg {
			if i == len(queryErrorArg)-1 {
				s += fmt.Sprintf(" %s.", query)
			} else {
				s += fmt.Sprintf(" %s", query)
			}
		}

		return queryValues, errors.New(s)
	}

	return queryValues, nil
}

// Utilizare cheie in context
func (c *Ctx) InitKeys() {
	c.keys = make(map[string]any)
}

func (c *Ctx) SetKey(key string, value any) {
	if c.keys == nil {
		c.InitKeys()
	}

	c.keys[key] = value
}

func (c *Ctx) GetKey(key string) (any, error) {
	if c.keys == nil {
		return nil, ErrKeysInexistent
	}

	if value, ok := c.keys[key]; ok {
		return value, nil
	}

	return nil, ErrKeyNotFound
}

func (c *Ctx) DeleteKey(key string) error {
	if c.keys == nil {
		return ErrKeysInexistent
	}

	if _, ok := c.keys[key]; ok {
		delete(c.keys, key)
		return nil
	}

	return ErrKeyNotFound
}

func (c *Ctx) GetKeysArray(keys ...string) ([]any, error) {
	var values []any
	var missingKeys []string

	for _, key := range keys {
		if value, ok := c.keys[key]; ok {
			values = append(values, value)
		} else {
			missingKeys = append(missingKeys, key)
		}
	}

	if len(missingKeys) > 0 {
		errMsg := "Next keys are non-existent:"
		for i, k := range missingKeys {
			if i == len(missingKeys)-1 {
				errMsg += fmt.Sprintf(" %s.", k)
			} else {
				errMsg += fmt.Sprintf(" %s", k)
			}
		}
		return values, errors.New(errMsg)
	}

	return values, nil
}

// func (c *Ctx)

func (c *Ctx) GetKeysMap() map[string]any {
	return c.keys
}

func (c *Ctx) GetStringKey(key string) (string, error) {
	value, err := c.GetKey(key)
	if err != nil {
		return "", err
	}

	if stringValue, ok := value.(string); ok {
		return stringValue, nil
	}
	return "", ErrInvalidType
}

func (c *Ctx) GetIntKey(key string) (int, error) {
	value, err := c.GetKey(key)
	if err != nil {
		return 0, err
	}

	if intValue, ok := value.(int); ok {
		return intValue, nil
	}
	return 0, ErrInvalidType
}

func (c *Ctx) GetInt64Key(key string) (int64, error) {
	value, err := c.GetKey(key)
	if err != nil {
		return int64(0), err
	}

	if intValue, ok := value.(int64); ok {
		return intValue, nil
	}
	return int64(0), ErrInvalidType
}

func (c *Ctx) GetInt32Key(key string) (int32, error) {
	value, err := c.GetKey(key)
	if err != nil {
		return int32(0), err
	}

	if intValue, ok := value.(int32); ok {
		return intValue, nil
	}
	return int32(0), ErrInvalidType
}

func (c *Ctx) GetFloat64Key(key string) (float64, error) {
	value, err := c.GetKey(key)
	if err != nil {
		return float64(0), err
	}
	if floatValue, ok := value.(float64); ok {
		return floatValue, nil
	}
	return float64(0), ErrInvalidType
}

func (c *Ctx) GetBoolKey(key string) (bool, error) {
	value, err := c.GetKey(key)
	if err != nil {
		return false, err
	}
	if boolValue, ok := value.(bool); ok {
		return boolValue, nil
	}
	return false, ErrInvalidType
}

func (c *Ctx) GetCookieValue(name string) (string, error) {
	cookie, err := c.r.Cookie(name)
	if err != nil {
		return "", ErrCookieNotFound
	}

	// https://golang.org = https%3A%2F%2Fgolang.org
	unescapedCookie := url.QueryEscape(cookie.Value)

	return unescapedCookie, nil
}

// Get all cookies from request
func (c *Ctx) Cookies() []*Cookie {
	cookies := c.r.Cookies()
	cookieList := make([]*Cookie, len(cookies))

	for i, cookie := range cookies {
		cookieList[i] = &Cookie{
			Name:       cookie.Name,
			Value:      cookie.Value,
			MaxAge:     cookie.MaxAge,
			RawExpires: cookie.RawExpires,
			Secure:     cookie.Secure,
			HttpOnly:   cookie.HttpOnly,
			SameSite:   SameSite(cookie.SameSite),
			Domain:     cookie.Domain,
		}
	}

	return cookieList
}

func (c *Ctx) GetCookie(name string) (*Cookie, error) {
	cookie, err := c.r.Cookie(name)
	if err != nil {
		return nil, ErrCookieNotFound
	}

	return &Cookie{
		Name:       cookie.Name,
		Value:      cookie.Value,
		MaxAge:     cookie.MaxAge,
		RawExpires: cookie.RawExpires,
		Secure:     cookie.Secure,
		HttpOnly:   cookie.HttpOnly,
		SameSite:   SameSite(cookie.SameSite),
		Domain:     cookie.Domain,
	}, nil
}

func (cookie *Cookie) SetSecure(secure bool) {
	cookie.Secure = secure
}

func (cookie *Cookie) SetHttpOnly(httpOnly bool) {
	cookie.HttpOnly = httpOnly
}

func (cookie *Cookie) SetSameSite(sameSite SameSite) {
	cookie.SameSite = sameSite
}

func (cookie *Cookie) SetDomain(domain string) {
	cookie.Domain = domain
}

func (cookie *Cookie) SetPath(path string) {
	cookie.Path = path
}

func (cookie *Cookie) SetMaxAge(maxAge int) {
	cookie.MaxAge = maxAge
}

// RawExpires nu este prioritar, ex: RawExpires: "Wed, 08 Jan 2025 12:00:00 GMT"
func (cookie *Cookie) SetRawExpires(rawExpires string) {
	cookie.RawExpires = rawExpires
}

func (c *Ctx) SetCookie(cookie *Cookie) error {
	if cookie.Name == "" || cookie.Value == "" {
		return ErrInvalidCookieParam
	}

	if cookie.Domain == "" {
		cookie.Domain = c.r.URL.Host
	}

	if cookie.Path == "" {
		cookie.Path = "/"
	}

	if cookie.MaxAge == 0 {
		cookie.MaxAge = 86400
	}

	if cookie.RawExpires == "" {
		cookie.RawExpires = "3600"
	}

	ck := &http.Cookie{
		Name:       cookie.Name,
		Value:      cookie.Value,
		MaxAge:     cookie.MaxAge,
		Secure:     cookie.Secure,
		RawExpires: cookie.RawExpires,
		Path:       cookie.Path,
		Domain:     cookie.Domain,
		HttpOnly:   cookie.HttpOnly,
		SameSite:   http.SameSite(cookie.SameSite),
	}

	http.SetCookie(c.w, ck)
	return nil
}

func (c *Ctx) GetProtoInfo() ProtoInfo {
	return ProtoInfo{
		Proto:      c.r.Proto,
		ProtoMajor: c.r.ProtoMajor,
		ProtoMinor: c.r.ProtoMinor,
	}
}

func (c *Ctx) ContentLength() int64 {
	return c.r.ContentLength
}

func (c *Ctx) TransferEncoding() []string {
	return c.r.TransferEncoding
}

func (c *Ctx) GetMethod() string {
	return c.r.Method
}

// returneaza headerul la requestul primit de server
func (c *Ctx) GetHeader(header string) (string, error) {
	headerValue := c.r.Header.Get(header)

	if headerValue == "" {
		return "", ErrHeaderNotFound
	}

	return headerValue, nil
}

func (c *Ctx) GetAllHeaders(headers ...string) (map[string]string, error) {
	result := make(map[string]string)

	for _, header := range headers {
		if value := c.r.Header.Get(header); value != "" {
			result[header] = value
		}
	}

	if len(result) == 0 {
		return nil, ErrHeaderNotFound
	}

	if len(result) != len(headers) {
		return result, ErrOneOrMoreHeadersNotFound
	}

	return result, nil
}

// seteaza headerul la response trimis de SERVER
func (c *Ctx) SetHeader(header, value string) {
	c.w.Header().Set(header, value)
}

func (c *Ctx) AddHeader(header, value string) {
	c.w.Header().Add(header, value)
	c.r.Header.Add(header, value)
}

func (c *Ctx) RemoveHeader(header string) {
	c.w.Header().Del(header)
}

func (c *Ctx) Body() ([]byte, error) {
	body, err := io.ReadAll(c.r.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read request body: %w", err)
	}

	if err := c.r.Body.Close(); err != nil {
		return nil, fmt.Errorf("failed to close request body: %w", err)
	}

	return body, nil
}

func (c *Ctx) BodyParsed() (J, error) {
	body, err := c.Body()
	if err != nil {
		return nil, fmt.Errorf("failed to read request body: %w", err)
	}

	bodyParsed := J{}
	err = json.Unmarshal(body, &bodyParsed)
	if err != nil {
		return nil, fmt.Errorf("failed to parse request body: %w", err)
	}
	return bodyParsed, nil
}

// to do: client - application/json => server - json - ok "json"
func (c *Ctx) Accept(headerServerAccept ...string) string {
	if len(headerServerAccept) == 0 {
		return ""
	}

	acceptHeader, err := c.GetHeader("Accept")
	if err != nil {
		return ""
	}

	headerClientAccept, err := getRequestHeaderValues(acceptHeader)
	if err != nil {
		return ""
	}

	sort.Slice(headerClientAccept, func(i, j int) bool {
		return headerClientAccept[i].q > headerClientAccept[j].q
	})

	for _, clientAccept := range headerClientAccept {
		for _, serverAccept := range headerServerAccept {
			svParts := strings.Split(serverAccept, ";")
			svMediaType := strings.TrimSpace(svParts[0])

			svParams := make(map[string]string)
			for i := 1; i < len(svParts); i++ {
				parts := strings.SplitN(strings.TrimSpace(svParts[i]), "=", 2)
				if len(parts) == 2 {
					svParams[strings.ToLower(parts[0])] = strings.ToLower(parts[1])
				}
			}

			// accepts case insensitive
			if strings.EqualFold(clientAccept.acceptHeaderValue, svMediaType) || clientAccept.acceptHeaderValue == "*/*" {
				if len(clientAccept.variants) > 0 {
					variantsSupported := true

					for _, variant := range clientAccept.variants {
						varParts := strings.SplitN(variant, "=", 2)
						if len(varParts) == 2 {
							paramName := varParts[0]

							if _, exists := svParams[paramName]; !exists {
								variantsSupported = false
								break
							}
						}
					}

					if variantsSupported {
						return serverAccept
					}
				} else {
					return serverAccept
				}
			}

			if clientAccept.acceptHeaderValue == "*/*" {
				return clientAccept.acceptHeaderValue
			}

			argsClientMediaType := strings.Split(clientAccept.acceptHeaderValue, "/")

			for i := 0; i < len(argsClientMediaType); i++ {
				argsClientMediaType[i] = strings.ToLower(argsClientMediaType[i])
				if argsClientMediaType[i] == svMediaType {
					return serverAccept
				}
				if argsClientMediaType[i] == "*" {
					return serverAccept
				}
			}
		}
	}
	return ""
}

func (c *Ctx) AcceptLanguage(headerServerAcceptLanguage ...string) string {
	clientLanguage, err := c.GetHeader("Accept-Language")
	if err != nil {
		return ""
	}

	clientLanguageExtract, err := getBaseHeader(clientLanguage)
	if err != nil {
		return ""
	}

	return c.matchAcceptValues(clientLanguageExtract, headerServerAcceptLanguage, "-")
}

func (c *Ctx) AcceptCharset(headerServerAcceptCharset ...string) string {
	clientCharset, err := c.GetHeader("Accept-Charset")
	if err != nil {
		return ""
	}

	clientCharsetExtract, err := getBaseHeader(clientCharset)
	if err != nil {
		return ""
	}

	return c.matchAcceptValues(clientCharsetExtract, headerServerAcceptCharset, "")
}

func (c *Ctx) AcceptEncoding(headerServerAcceptEncoding ...string) string {
	clientEncoding, err := c.GetHeader("Accept-Encoding")
	if err != nil {
		return ""
	}

	clientEncodingExtract, err := getBaseHeader(clientEncoding)
	if err != nil {
		return ""
	}

	return c.matchAcceptValues(clientEncodingExtract, headerServerAcceptEncoding, "")
}

func (c *Ctx) matchAcceptValues(
	clientValues []acceptSpecial,
	serverValues []string,
	separator string) string {
	for _, clientValue := range clientValues {
		for _, serverValue := range serverValues {
			if clientValue.acceptHeaderValue == "*" {
				return serverValue
			}

			if strings.EqualFold(clientValue.acceptHeaderValue, serverValue) {
				return serverValue
			}

			if separator != "" {
				serverParts := strings.SplitN(serverValue, separator, 2)
				if len(serverParts) == 2 {
					basePart := strings.ToLower(serverParts[0])
					clientBase := strings.ToLower(clientValue.acceptHeaderValue)

					if clientBase == basePart {
						return serverValue
					}
				}
			}
		}
	}

	return ""
}

func getBaseHeader(clientCharset string) ([]acceptSpecial, error) {
	if clientCharset == "" {
		return nil, ErrHeaderNotFound
	}

	clientCharsetSplit := strings.Split(clientCharset, ",")

	if len(clientCharsetSplit) == 0 {
		return nil, ErrHeaderNotFound
	}

	acceptedValues := make([]acceptSpecial, 0, len(clientCharsetSplit))

	for _, clientCharsetSplitValue := range clientCharsetSplit {
		clientValues := strings.Split(strings.TrimSpace(clientCharsetSplitValue), ";")
		q := 1.0

		if len(clientValues) > 1 {
			if strings.HasPrefix(clientValues[1], "q=") {
				qParts := strings.Split(clientValues[1], "=")
				if len(qParts) >= 2 {
					qValue, err := strconv.ParseFloat(strings.TrimSpace(qParts[1]), 64)
					if err != nil {
						return nil, err
					}
					q = qValue
				}
			}
		}

		acceptedValues = append(acceptedValues, acceptSpecial{
			acceptHeaderValue: strings.TrimSpace(clientValues[0]),
			q:                 q,
		})
	}

	sort.Slice(acceptedValues, func(i, j int) bool {
		return acceptedValues[i].q > acceptedValues[j].q
	})

	return acceptedValues, nil
}

func getRequestHeaderValues(accept string) ([]acceptHeader, error) {

	acceptedHeaderSplit := strings.Split(accept, ",")
	if len(acceptedHeaderSplit) == 0 {
		return nil, ErrHeaderNotFound
	}

	acceptedValues := make([]acceptHeader, 0, len(acceptedHeaderSplit))
	if accept == "" {
		return nil, ErrHeaderNotFound
	}

	for _, acceptHeaderValueRaw := range acceptedHeaderSplit {

		var variants []string

		if acceptHeaderValueRaw == "" {
			continue
		}
		splitValues := strings.Split(strings.TrimSpace(acceptHeaderValueRaw), ";")

		q := 1.0

		if len(splitValues) > 1 {
			for i := 1; i < len(splitValues); i++ {
				param := strings.TrimSpace(splitValues[i])
				if strings.HasPrefix(param, "q=") {
					qParts := strings.Split(param, "=")
					if len(qParts) >= 2 {
						q, _ = strconv.ParseFloat(qParts[1], 64)
					}
				} else {
					variants = append(variants, param)
				}
			}
		}

		if q > 0 {
			acceptedValues = append(acceptedValues, acceptHeader{
				acceptHeaderValue: strings.TrimSpace(splitValues[0]),
				q:                 q,
				variants:          variants,
			})
		}
	}

	return acceptedValues, nil
}

// ParseForm parses the form data from the request body
// if not used it will be parsed automatically
func (c *Ctx) ParseForm() error {
	contentType, err := c.GetHeader("Content-Type")
	if err != nil {
		return fmt.Errorf("failed to get content type: %w", err)
	}

	if c.formParsed {
		return fmt.Errorf("form already parsed")
	}

	if strings.Contains(contentType, "application/x-www-form-urlencoded") {
		err = c.r.ParseForm()
		if err != nil {
			return fmt.Errorf("failed to parse form: %w", err)
		}

		c.formParsed = true
		return nil
	} else if strings.Contains(contentType, "multipart/form-data") ||
		strings.Contains(contentType, "application/octet-stream") {
		err = c.r.ParseMultipartForm(c.maxMemory)
		if err != nil {
			return fmt.Errorf("failed to parse multipart form: %w", err)
		}
		c.formParsed = true
		return nil
	}

	return nil
}

// form works only for application/x-www-form-urlencoded and form-data without binary values
func (c *Ctx) Form() (url.Values, error) {
	if !c.formParsed {
		err := c.ParseForm()
		if err != nil {
			return nil, fmt.Errorf("failed to parse form: %w", err)
		}
	}

	form := c.r.Form
	if form == nil {
		return nil, fmt.Errorf("form is empty")
	}

	return form, nil
}

// should add MultipartForm also

func (c *Ctx) FormFile(key string) (TreeFile, error) {

	treeNil := TreeFile{
		MultipartFile:       nil,
		MultipartFileHeader: nil,
	}

	if !c.formParsed {
		err := c.ParseForm()
		if err != nil {
			return treeNil, fmt.Errorf("failed to parse form: %w", err)
		}
	}

	if c.r.MultipartForm == nil {
		contentType := c.r.Header.Get("Content-Type")
		mediaType, _, _ := mime.ParseMediaType(contentType)
		if mediaType != "multipart/form-data" {
			return treeNil, fmt.Errorf("form is not multipart/form-data")
		}

		return treeNil, fmt.Errorf("multipart form is empty")
	}

	file, fileheader, err := c.r.FormFile(key)
	if err != nil {
		if errors.Is(err, http.ErrMissingFile) {
			return treeNil, fmt.Errorf("no file found for key: %s", key)
		}
		return treeNil, fmt.Errorf("failed to get form file: %w", err)
	}

	return TreeFile{
		MultipartFile:       file,
		MultipartFileHeader: fileheader,
	}, nil
}

type BinaryFile struct {
	MediaType         string
	DispositionParams map[string]string
	Data              []byte
}

func (c *Ctx) GetBodyAsBinary() (b BinaryFile, err error) {
	contentType, err := c.GetHeader("Content-Type")
	binaryFileNil := BinaryFile{
		MediaType:         "",
		DispositionParams: nil,
		Data:              nil,
	}
	if err != nil {
		return binaryFileNil, fmt.Errorf("failed to get content type: %w", err)
	}

	mediaType, _, _ := mime.ParseMediaType(contentType)
	if mediaType == "multipart/form-data" || mediaType == "application/x-www-form-urlencoded" {
		return binaryFileNil, fmt.Errorf("form is not binary")
	}

	data, err := io.ReadAll(c.r.Body)
	if err != nil {
		return binaryFileNil, fmt.Errorf("failed to read request body: %w", err)
	}

	err = c.r.Body.Close()
	if err != nil {
		return binaryFileNil, fmt.Errorf("failed to close request body: %w", err)
	}

	var dispositionParams map[string]string
	disposition, err := c.GetHeader("Content-Disposition")

	if err == nil && disposition != "" {
		_, params, parseErr := mime.ParseMediaType(disposition)
		if parseErr != nil {
			return binaryFileNil, fmt.Errorf("failed to parse content disposition: %w", parseErr)
		} else {
			dispositionParams = params
		}
	} else if err != nil && !errors.Is(err, ErrHeaderNotFound) {
		return binaryFileNil, fmt.Errorf("failed to get content disposition: %w", err)
	}

	return BinaryFile{
		MediaType:         mediaType,
		DispositionParams: dispositionParams,
		Data:              data,
	}, nil
}

func (c *Ctx) FormFiles(keys ...string) (map[string]TreeFile, error) {
	if len(keys) == 0 {
		return nil, fmt.Errorf("no keys provided")
	}

	if !c.formParsed {
		err := c.ParseForm()
		if err != nil {
			return nil, fmt.Errorf("failed to parse form: %w", err)
		}
	}

	if c.r.MultipartForm == nil {
		contentType, err := c.GetHeader("Content-Type")
		if err != nil {
			return nil, fmt.Errorf("failed to get content type: %w", err)
		}

		mediaType, _, _ := mime.ParseMediaType(contentType)
		if mediaType != "multipart/form-data" {
			return nil, fmt.Errorf("form is not multipart/form-data")
		}
		return nil, fmt.Errorf("multipart form is empty")
	}

	files := make(map[string]TreeFile)

	if len(keys) == 1 && keys[0] == "all" {
		for key, fileHeaders := range c.r.MultipartForm.File {
			for _, fileHeader := range fileHeaders {
				file, err := fileHeader.Open()
				if err != nil {
					fmt.Printf("[ERROR] %s.\n", err.Error())
					continue
				}
				files[key] = TreeFile{
					MultipartFile:       file,
					MultipartFileHeader: fileHeader,
				}
			}
		}
	} else {
		for _, key := range keys {
			file, fileHeader, err := c.r.FormFile(key)
			if err != nil {
				fmt.Printf("[ERROR] %s.\n", err.Error())
				continue
			}
			files[key] = TreeFile{
				MultipartFile:       file,
				MultipartFileHeader: fileHeader,
			}
		}
	}

	if len(files) == 0 && !(len(keys) == 1 && keys[0] == "all") {
		return nil, fmt.Errorf("no files found for the specified keys")
	}

	return files, nil
}

func (c *Ctx) FormValue(key string) (string, error) {
	if !c.formParsed {
		fmt.Printf("[WARNING] Form not parsed yet. Parsing now...\n")
		err := c.ParseForm()
		if err != nil {
			return "", fmt.Errorf("failed to parse form: %w", err)
		}
	}

	value := c.r.FormValue(key)
	if value != "" {
		return value, nil
	}

	return "", fmt.Errorf("form value not found or empty for key: %s", key)
}

// de adaugat  valoarea pentru FormFiles
func (c *Ctx) FormValues(key ...string) (map[string]string, error) {
	if len(key) == 0 {
		return nil, fmt.Errorf("no keys provided")
	}

	if !c.formParsed {
		err := c.ParseForm()
		if err != nil {
			return nil, fmt.Errorf("failed to parse form: %w", err)
		}
	}

	formMap := make(map[string]string)
	for _, k := range key {
		value, err := c.FormValue(k)
		if err != nil {
			fmt.Printf("[ERROR] %s.\n", err.Error())
			continue
		}
		formMap[k] = value
	}

	if len(formMap) == 0 {
		return nil, fmt.Errorf("no form values found")
	}
	return formMap, nil
}

// SetMaxMemParsed sets the maximum memory to be used for parsing form data
// in bytes. The default value is 10MB.
//
// Note: This function is only relevant for multipart/form-data requests.
// Use before calling ParseForm()
func (c *Ctx) SetMaxMemParsed(maxMemory int64) {
	c.maxMemory = maxMemory
}

func (c *Ctx) GetMaxMemParsed() int64 {
	return c.maxMemory
}

func (c *Ctx) BindJSON(obj any) error {
	err := binding.JSON.Bind(c.r, obj)
	if err != nil {
		return err
	}
	return nil
}

func (c *Ctx) BindXML(obj any) error {
	err := binding.XML.Bind(c.r, obj)
	if err != nil {
		return err
	}
	return nil
}

func (c *Ctx) BindYAML(obj any) error {
	err := binding.YAML.Bind(c.r, obj)
	if err != nil {
		return err
	}
	return nil
}

func (c *Ctx) BindTOML(obj any) error {
	err := binding.TOML.Bind(c.r, obj)
	if err != nil {
		return err
	}
	return nil
}

func (c *Ctx) BindPlaintext(obj any) error {
	err := binding.Text.Bind(c.r, obj)
	if err != nil {
		return err
	}
	return nil
}

func (c *Ctx) BindURI(obj any) error {
	err := binding.URI.BindURI(c.params, obj)
	if err != nil {
		return err
	}
	return nil
}

func (c *Ctx) BindQuery(obj any) error {
	err := binding.Query.BindQuery(c.r, obj)
	if err != nil {
		return err
	}
	return nil
}

func (c *Ctx) BindForm(obj any) error {
	err := binding.Form.Bind(c.r, obj)
	if err != nil {
		return err
	}
	return nil
}

func (c *Ctx) GetRequest() *http.Request {
	return c.r
}

func (c *Ctx) Render(code int, r render.Render) error {
	c.w.WriteHeader(code)
	return r.Render(c.w)
}

func (c *Ctx) RenderJSON(code int, r render.JSONRender) error {
	c.w.WriteHeader(code)
	return r.Render(c.w)
}

func (c *Ctx) Path() string {
	if c.r == nil || c.r.URL == nil {
		return ""
	}
	return c.r.URL.Path
}
