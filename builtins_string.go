package schematics

import (
	"errors"
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"unicode/utf8"
)

var (
	errNotString = errors.New("value is not a string")

	reEmail    = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
	reURL      = regexp.MustCompile(`^(https?)://[a-zA-Z0-9.\-]+(\.[a-zA-Z]{2,})?(:\d+)?(/.*)?$`)
	reUUID     = regexp.MustCompile(`^[a-fA-F0-9]{8}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{12}$`)
	reNonAlnum = regexp.MustCompile(`[^a-zA-Z0-9]`)
	reUpper    = regexp.MustCompile(`[A-Z]`)
	reLower    = regexp.MustCompile(`[a-z]`)
	reDigit    = regexp.MustCompile(`[0-9]`)
)

func ruleIsString(v any, _ Args, _ *Context) error {
	if _, ok := v.(string); !ok {
		return errNotString
	}
	return nil
}

func ruleNotEmpty(v any, _ Args, _ *Context) error {
	s, ok := asString(v)
	if !ok {
		return errNotString
	}
	if strings.TrimSpace(s) == "" {
		return errors.New("must not be empty")
	}
	return nil
}

func ruleEmail(v any, _ Args, _ *Context) error {
	s, ok := asString(v)
	if !ok {
		return errNotString
	}
	if !reEmail.MatchString(s) {
		return fmt.Errorf("%q is not a valid email address", s)
	}
	return nil
}

func ruleMaxLength(v any, args Args, _ *Context) error {
	s, ok := asString(v)
	if !ok {
		return errNotString
	}
	max, err := args.Int("max")
	if err != nil {
		return err
	}
	if utf8.RuneCountInString(s) > max {
		return fmt.Errorf("must be at most %d characters", max)
	}
	return nil
}

func ruleMinLength(v any, args Args, _ *Context) error {
	s, ok := asString(v)
	if !ok {
		return errNotString
	}
	min, err := args.Int("min")
	if err != nil {
		return err
	}
	if utf8.RuneCountInString(s) < min {
		return fmt.Errorf("must be at least %d characters", min)
	}
	return nil
}

func ruleLengthBetween(v any, args Args, _ *Context) error {
	s, ok := asString(v)
	if !ok {
		return errNotString
	}
	min, err := args.Int("min")
	if err != nil {
		return err
	}
	max, err := args.Int("max")
	if err != nil {
		return err
	}
	n := utf8.RuneCountInString(s)
	if n < min || n > max {
		return fmt.Errorf("must be between %d and %d characters", min, max)
	}
	return nil
}

func ruleNoSpecialChars(v any, _ Args, _ *Context) error {
	s, ok := asString(v)
	if !ok {
		return errNotString
	}
	if reNonAlnum.MatchString(s) {
		return errors.New("must not contain special characters")
	}
	return nil
}

func ruleHasSpecialChars(v any, _ Args, _ *Context) error {
	s, ok := asString(v)
	if !ok {
		return errNotString
	}
	if !reNonAlnum.MatchString(s) {
		return errors.New("must contain at least one special character")
	}
	return nil
}

func ruleHasUpper(v any, _ Args, _ *Context) error {
	s, ok := asString(v)
	if !ok {
		return errNotString
	}
	if !reUpper.MatchString(s) {
		return errors.New("must contain at least one uppercase letter")
	}
	return nil
}

func ruleHasLower(v any, _ Args, _ *Context) error {
	s, ok := asString(v)
	if !ok {
		return errNotString
	}
	if !reLower.MatchString(s) {
		return errors.New("must contain at least one lowercase letter")
	}
	return nil
}

func ruleHasDigit(v any, _ Args, _ *Context) error {
	s, ok := asString(v)
	if !ok {
		return errNotString
	}
	if !reDigit.MatchString(s) {
		return errors.New("must contain at least one digit")
	}
	return nil
}

func ruleIsURL(v any, _ Args, _ *Context) error {
	s, ok := asString(v)
	if !ok {
		return errNotString
	}
	if !reURL.MatchString(s) {
		return fmt.Errorf("%q is not a valid URL", s)
	}
	return nil
}

func ruleNotURL(v any, _ Args, _ *Context) error {
	s, ok := asString(v)
	if !ok {
		return errNotString
	}
	if reURL.MatchString(s) {
		return fmt.Errorf("%q must not be a URL", s)
	}
	return nil
}

func ruleURLHasHost(v any, args Args, _ *Context) error {
	s, ok := asString(v)
	if !ok {
		return errNotString
	}
	host, err := args.String("host")
	if err != nil {
		return err
	}
	u, perr := url.Parse(s)
	if perr != nil {
		return fmt.Errorf("%q is not a parseable URL", s)
	}
	if !strings.HasSuffix(u.Hostname(), host) {
		return fmt.Errorf("host %q does not match %q", u.Hostname(), host)
	}
	return nil
}

func ruleURLHasQuery(v any, args Args, _ *Context) error {
	s, ok := asString(v)
	if !ok {
		return errNotString
	}
	params, err := args.Strings("params")
	if err != nil {
		// also accept a single comma-separated string
		if raw, serr := args.String("params"); serr == nil {
			params = splitAndTrim(raw)
		} else {
			return err
		}
	}
	u, perr := url.Parse(s)
	if perr != nil {
		return fmt.Errorf("%q is not a parseable URL", s)
	}
	q := u.Query()
	for _, p := range params {
		if _, ok := q[strings.TrimSpace(p)]; !ok {
			return fmt.Errorf("missing query parameter %q", p)
		}
	}
	return nil
}

func ruleIsHTTPS(v any, _ Args, _ *Context) error {
	s, ok := asString(v)
	if !ok {
		return errNotString
	}
	u, err := url.Parse(s)
	if err != nil || u.Scheme != "https" {
		return fmt.Errorf("%q is not an https URL", s)
	}
	return nil
}

func ruleIsUUID(v any, _ Args, _ *Context) error {
	s, ok := asString(v)
	if !ok {
		return errNotString
	}
	if !reUUID.MatchString(s) {
		return fmt.Errorf("%q is not a valid UUID", s)
	}
	return nil
}

func ruleEquals(v any, args Args, _ *Context) error {
	s, ok := asString(v)
	if !ok {
		return errNotString
	}
	want, err := args.String("value")
	if err != nil {
		return err
	}
	if s != want {
		return fmt.Errorf("must equal %q", want)
	}
	return nil
}

func ruleInOptions(v any, args Args, _ *Context) error {
	s, ok := asString(v)
	if !ok {
		return errNotString
	}
	opts, err := args.Strings("options")
	if err != nil {
		return err
	}
	for _, o := range opts {
		if o == s {
			return nil
		}
	}
	return fmt.Errorf("%q is not one of the allowed values", s)
}

func ruleMatchRegex(v any, args Args, _ *Context) error {
	s, ok := asString(v)
	if !ok {
		return errNotString
	}
	pattern, err := args.String("pattern")
	if err != nil {
		return err
	}
	re, cerr := regexp.Compile(pattern)
	if cerr != nil {
		return fmt.Errorf("invalid regex %q: %w", pattern, cerr)
	}
	if !re.MatchString(s) {
		return fmt.Errorf("must match %q", pattern)
	}
	return nil
}

// ruleLike implements SQL-style LIKE matching (% is any run, _ is any single
// character); all other characters are matched literally.
func ruleLike(v any, args Args, _ *Context) error {
	s, ok := asString(v)
	if !ok {
		return errNotString
	}
	pattern, err := args.String("pattern")
	if err != nil {
		return err
	}
	quoted := regexp.QuoteMeta(pattern)
	quoted = strings.ReplaceAll(quoted, "%", ".*")
	quoted = strings.ReplaceAll(quoted, "_", ".")
	re := regexp.MustCompile("^" + quoted + "$")
	if !re.MatchString(s) {
		return fmt.Errorf("%q is not like %q", s, pattern)
	}
	return nil
}

func splitAndTrim(s string) []string {
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if t := strings.TrimSpace(p); t != "" {
			out = append(out, t)
		}
	}
	return out
}
