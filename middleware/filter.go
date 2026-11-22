package middleware

import (
	"fmt"
	"net/http"
	"regexp"

	"nelia/configuration"

	"github.com/esequiel378/hl7"
)

type IngoingFilter struct {
	ValidConfig          bool
	HeaderRules          map[string]FilteringRule
	URLParameterRules    map[string]FilteringRule
	ForbiddenHL7Segments map[string]string
}

// All rules are blacklists (for simplicity, adding whitelist would be longer but not harder)
// Key is either the header, URL parameter, or segment name to filter on
// Match is a regexp. Any value for the field that match will lead the request to be denied
// Name is a human-understandable name for the rule, so we can serve the reason of the denial
type FilteringRule struct {
	Key   string
	Match *regexp.Regexp
	Name  string
}

// If the configuration is incorrect we reject everything, we don't use the old config
// This is a way to make configuration errors never be silent
func (i *IngoingFilter) RejectWithoutValidConfig(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !i.ValidConfig {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Gateway configuration error"))
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (i *IngoingFilter) ApplyHeaderRules(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		for h, values := range r.Header {
			if rule, ok := i.HeaderRules[h]; ok {
				for _, value := range values {
					if rule.Match.Match([]byte(value)) {
						w.WriteHeader(http.StatusForbidden)
						w.Write([]byte(fmt.Sprintf("Request refused due to header rule %s", rule.Name)))
						return
					}
				}
			}
		}
		next.ServeHTTP(w, r)
	})
}

func (i *IngoingFilter) ApplyParameterRules(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		for p, values := range r.URL.Query() {
			if rule, ok := i.URLParameterRules[p]; ok {
				for _, value := range values {
					if rule.Match.Match([]byte(value)) {
						w.WriteHeader(http.StatusForbidden)
						w.Write([]byte(fmt.Sprintf("Request refused due to URL parameter rule %s", rule.Name)))
						return
					}
				}
			}
		}
		next.ServeHTTP(w, r)
	})
}

// This is not a middleware because request bodies can only be read once, so it need to be read by the final handler, not a middleware
// (workarounds exists but would be overall messier than applying this function in the handler)
func (i *IngoingFilter) ApplyHL7Rules(message *hl7.GenericMessage) error {
	for _, segment := range message.Segments {
		if name, ok := i.ForbiddenHL7Segments[segment.Name]; ok {
			return fmt.Errorf("Request refused due to HL7 segment rule %s", name)
		}
	}
	return nil
}

// These functions let us refuse all requests while configuration is being updated
func (i *IngoingFilter) Disable() {
	i.ValidConfig = false
}

func (i *IngoingFilter) Enable() {
	i.ValidConfig = true
}

// Configuration is made with regular slices, so we need to convert them to map
func (i *IngoingFilter) UpdateWithConfig(conf configuration.Configuration) error {
	newHeaderRules := make(map[string]FilteringRule)
	for _, rule := range conf.HeaderRules {
		regexp, err := regexp.Compile(rule.UnparsedMatcher)
		if err != nil {
			return fmt.Errorf("Error parsing regexp %s for header rule %s: %w", rule.UnparsedMatcher, rule.Name, err)
		}
		newRule := FilteringRule{
			Key:   rule.Key,
			Match: regexp,
			Name:  rule.Name,
		}
		newHeaderRules[rule.Key] = newRule
	}
	i.HeaderRules = newHeaderRules

	newParameterRules := make(map[string]FilteringRule)
	for _, rule := range conf.ParameterRules {
		regexp, err := regexp.Compile(rule.UnparsedMatcher)
		if err != nil {
			return fmt.Errorf("Error parsing regexp %s for parameter rule %s: %w", rule.UnparsedMatcher, rule.Name, err)
		}
		newRule := FilteringRule{
			Key:   rule.Key,
			Match: regexp,
			Name:  rule.Name,
		}
		newParameterRules[rule.Key] = newRule
	}
	i.URLParameterRules = newParameterRules

	newHL7Rules := make(map[string]string)
	for _, rule := range conf.SegmentRules {
		newHL7Rules[rule.Key] = rule.Name
	}
	i.ForbiddenHL7Segments = newHL7Rules

	return nil
}
