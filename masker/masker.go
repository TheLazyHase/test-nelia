package masker

import (
	"fmt"
	"strings"

	"nelia/configuration"

	"github.com/esequiel378/hl7"
)

const separator = `|`

// Map are faster than slices to find an item in so we store rules in form of maps
type Masker struct {
	Rules map[string]MaskRule
}

type MaskRule struct {
	Segment    string
	Fields     map[int]struct{}
	Everything bool
	Name       string
}

// This is part masking, part simplified marshaller
// It don't concern itself with components and repetitions and whatnot because the field.Value contain the raw form of these variations
func (m *Masker) GetMaskedVersion(message *hl7.GenericMessage) (string, error) {
	marshaled := ""
	for _, segment := range message.Segments {
		newSegment := []string{segment.Name}
		if rule, ok := m.Rules[segment.Name]; ok {
			for i := range segment.Fields {
				if _, ok := rule.Fields[i]; ok || rule.Everything {
					segment.Fields[i].Value = "U"
				}
			}
		}
		for _, field := range segment.Fields {
			newSegment = append(newSegment, field.Value)
		}
		line := strings.Join(newSegment, separator) + "\n"

		//The library is doing something strange with MSH lines (which are Message Header lines)
		//I don't have the time to re-analyze the code inside to understand why it's doing that, so here a quickfix
		if line[0:6] == `MSH|||` {
			line = `MSH|` + line[6:]
		}
		marshaled = marshaled + line
	}
	return strings.Trim(marshaled, "\n"), nil
}

// Configuration is made with regular slices, so we need to convert them to map
func (m *Masker) UpdateWithConfig(conf configuration.Configuration) error {
	newRules := make(map[string]MaskRule)
	for _, rule := range conf.MaskingRules {
		newRule := MaskRule{
			Segment:    rule.Key,
			Fields:     make(map[int]struct{}),
			Everything: rule.Everything,
			Name:       rule.Name,
		}
		// The +1 is because golang array are 0 based, and HL7 field are 1-based
		// So field 1 is at index 0
		for _, fieldID := range rule.Fields {
			if fieldID < 1 {
				return fmt.Errorf("Invalid field ID, must be at least 1: %d", fieldID)
			}
			newRule.Fields[fieldID-1] = struct{}{}
		}
		newRules[rule.Key] = newRule
	}
	//We hardcode a rule about identities
	newRules[`PID`] = MaskRule{
		Segment:    `PID`,
		Fields:     make(map[int]struct{}),
		Everything: true,
		Name:       "Remove patient identity",
	}
	m.Rules = newRules
	return nil
}
