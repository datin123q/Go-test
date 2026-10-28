package adebis

import (
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"
)

const (
	EbisAllowedOriginTest = "https://bishamon.ebis.ne.jp"
	EbisKeyHeaderMessage  = "X-Ebis-Set-Cookie-Message"
	EbisKeyHeaderTest     = "X-Ebis-Set-Cookie-Test"
	EbisCookieMaxAge      = 63072000
)

var ebisAllowedCookies = []string{"_ebtd", "_ebcv", "tracking_reject"}

type EbisCookie struct {
	req *http.Request
	res http.ResponseWriter
}

func NewSetEbisCookie(req *http.Request, res http.ResponseWriter) *EbisCookie {
	return &EbisCookie{
		req: req,
		res: res,
	}
}

func (s *EbisCookie) SetEbisCookie() {
	headers := map[string]string{"Content-Type": "text/plain"}

	defer func() {
		if r := recover(); r != nil {
			s.sendErrorHeader(fmt.Sprintf("Error: %v", r))
		}
	}()

	urlParts, err := url.Parse(s.req.URL.String())
	if err != nil {
		s.sendErrorHeader(fmt.Sprintf("URL parse error: %v", err))
		return
	}

	queryParams := urlParts.Query()
	domain := queryParams.Get("domain")
	name := queryParams.Get("name")
	value := queryParams.Get("value")
	test := queryParams.Get("test")
	checkCookie := false

	if value == "" {
		value = s.getCookie(name)
		checkCookie = true
	}

	invalid := s.validateParameters(domain, name, value, checkCookie)
	if invalid != "" {
		s.sendErrorHeader(invalid)
		return
	}

	if test != "" && value != "" {
		name = name + "_test"
		headers = s.getTestHeader(headers, s.getTestMessage(domain, name, value))
	}

	s.setCookie(domain, name, value)

	for key, val := range headers {
		s.res.Header().Set(key, val)
	}
	s.res.WriteHeader(http.StatusOK)
}

func (s *EbisCookie) validateParameters(domain, name, value string, checkCookie bool) string {
	if domain == "" || s.containsInvalidChars(domain) {
		return "Missing or invalid domain parameter"
	}

	if name == "" || !s.isAllowedCookie(name) {
		return "Missing or invalid name parameter"
	}

	if value == "" || s.containsInvalidChars(value) {
		if checkCookie {
			return "Cookie not found"
		}
		return "Invalid value parameter"
	}

	return ""
}

func (s *EbisCookie) isAllowedCookie(name string) bool {
	for _, allowed := range ebisAllowedCookies {
		if allowed == name {
			return true
		}
	}
	return false
}

func (s *EbisCookie) containsInvalidChars(str string) bool {
	matched, _ := regexp.MatchString(`[\n\r;]`, str)
	return matched
}

func (s *EbisCookie) sendErrorHeader(invalid string) {
	s.res.Header().Set(EbisKeyHeaderMessage, invalid)
	s.res.WriteHeader(http.StatusBadRequest)
}

func (s *EbisCookie) getTestHeader(headers map[string]string, message string) map[string]string {
	headers["Access-Control-Allow-Origin"] = EbisAllowedOriginTest
	headers["Access-Control-Expose-Headers"] = EbisKeyHeaderTest
	headers[EbisKeyHeaderTest] = message
	return headers
}

func (s *EbisCookie) getTestMessage(domain, name, value string) string {
	message := []string{
		fmt.Sprintf("%s=%s", name, value),
		fmt.Sprintf("Domain=.%s", domain),
		fmt.Sprintf("Max-Age=%d", EbisCookieMaxAge),
		"Path=/;",
	}
	return strings.Join(message, "; ")
}

func (s *EbisCookie) getCookie(key string) string {
	cookieHeader := s.req.Header.Get("Cookie")
	if cookieHeader == "" {
		return ""
	}

	cookies := strings.Split(cookieHeader, ";")
	for _, cookie := range cookies {
		cookie = strings.TrimSpace(cookie)
		if strings.HasPrefix(cookie, key+"=") {
			value := strings.TrimPrefix(cookie, key+"=")
			unescaped, err := url.QueryUnescape(value)
			if err != nil {
				return value
			}
			return unescaped
		}
	}
	return ""
}

func (s *EbisCookie) setCookie(domain, name, value string) {
	escapedName := url.QueryEscape(name)
	cookieHeader := []string{
		fmt.Sprintf("%s=%s", escapedName, value),
		fmt.Sprintf("Domain=.%s", domain),
		fmt.Sprintf("Max-Age=%d", EbisCookieMaxAge),
		"Path=/",
		"",
	}
	cookieHeaderString := strings.Join(cookieHeader, "; ")
	s.res.Header().Set("Set-Cookie", cookieHeaderString)
}
