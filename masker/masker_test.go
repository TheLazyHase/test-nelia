package masker_test

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"nelia/masker"

	"github.com/esequiel378/hl7"
)

func Test_GetMaskedVersion(t *testing.T) {
	t.Run("No rules", func(t *testing.T) {
		cases := []string{
			"example1_standard.hl7",
			"example2_ADT^A08.hl7",
			"example3_SIU^S14.hl7",
			"example4_numeric_with_notes.hl7",
			"example5_ORU^R03.hl7",
		}
		for _, file := range cases {
			initial, err := os.ReadFile("../examples_hf7/" + file)
			if err != nil {
				t.Fatalf("message failed to read : %s", err.Error())
			}

			mask := masker.Masker{
				Rules: make(map[string]masker.MaskRule),
			}
			parsedMessage, err := hl7.ParseGeneric(initial)
			if err != nil {
				t.Fatalf("message failed to parse : %s", err.Error())
			}
			//We need to mask fields since it's a HL7 result
			maskedMessage, err := mask.GetMaskedVersion(parsedMessage)
			if err != nil {
				t.Fatalf("error on masking : %s", err.Error())
			}
			inputTrimed := strings.Trim(string(initial), "\n")
			outputTrimed := strings.Trim(maskedMessage, "\n")
			if inputTrimed != outputTrimed {
				// We write in tmp the output to make debugging easier
				os.WriteFile("../tmp/old", []byte(inputTrimed), 0666)
				os.WriteFile("../tmp/new", []byte(outputTrimed), 0666)
				t.Fatalf("Erroneously changed output without rules : file %s", file)
			}
			fmt.Println(fmt.Sprintf("No-op test passed for %s", file))
		}
	})

	t.Run("Rule on every field of PID", func(t *testing.T) {
		cases := []string{
			"example1_standard.hl7",
			"example2_ADT^A08.hl7",
			"example3_SIU^S14.hl7",
			"example4_numeric_with_notes.hl7",
			"example5_ORU^R03.hl7",
		}
		for _, file := range cases {
			initial, err := os.ReadFile("../examples_hf7/" + file)
			if err != nil {
				t.Fatalf("message failed to read : %s", err.Error())
			}

			rules := make(map[string]masker.MaskRule)
			rules[`PID`] = masker.MaskRule{
				Segment:    `PID`,
				Fields:     make(map[int]struct{}),
				Everything: true,
				Name:       "Remove patient identity",
			}

			mask := masker.Masker{
				Rules: rules,
			}

			parsedMessage, err := hl7.ParseGeneric(initial)
			if err != nil {
				t.Fatalf("message failed to parse : %s", err.Error())
			}
			//We need to mask fields since it's a HL7 result
			maskedMessage, err := mask.GetMaskedVersion(parsedMessage)
			if err != nil {
				t.Fatalf("error on masking : %s", err.Error())
			}

			expected, err := os.ReadFile("test_results/" + file)
			if err != nil {
				t.Fatalf("error on reading expected : %s", err.Error())
			}
			expectedTrimed := strings.Trim(string(expected), "\n")
			outputTrimed := strings.Trim(maskedMessage, "\n")
			if expectedTrimed != outputTrimed {
				// We write in tmp the output to make debugging easier
				os.WriteFile("../tmp/old", []byte(expectedTrimed), 0666)
				os.WriteFile("../tmp/new", []byte(outputTrimed), 0666)
				t.Fatalf("Erroneously changed output with everything rule on PID : file %s", file)
			}
			fmt.Println(fmt.Sprintf("No-op test passed for %s", file))
		}
	})

	//We should add more tests but there's only 4 hours and the rest is a lot of busywork (trying different types of rules, etc)
}
