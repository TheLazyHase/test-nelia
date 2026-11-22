package main

import (
	"encoding/json"
	"fmt"
	"github.com/esequiel378/hl7"
	"log"
)

func main() {
	data := []byte(`PID|||PAT00000||SMITH^JANE||19900101|F
OBX|1|NM|GLU^Glycémie||1.05|g/L|0.7-1.10|N|||F`)

	msg, err := hl7.ParseGeneric(data)
	if err != nil {
		log.Fatal(err)
	}

	// Access segments, fields, components, and repetitions programmatically
	for _, seg := range msg.Segments {
		fmt.Printf("Segment: %s\n", seg.Name)
		for _, field := range seg.Fields {
			fmt.Printf("  %s-%d: %s\n", seg.Name, field.Index, field.Value)
		}
	}

	jsonData, _ := json.MarshalIndent(msg, "", "  ")
	fmt.Println(string(jsonData))

}
