package integration_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"nelia/configuration"
	"nelia/handler"
	"nelia/masker"
	"nelia/middleware"
)

type integrationTest struct {
	Input         string
	Output        string
	Configuration string
}

func Test_GetMaskedVersion(t *testing.T) {
	t.Run("nominal case", func(t *testing.T) {
		cases := []integrationTest{
			//TODO: add more test... but time is running out.
			{
				Input:         "examples_hf7/example1_standard.hl7",
				Output:        "integration_test_results/example1_standard.hl7",
				Configuration: "conf.test.json",
			},
			{
				Input:         "examples_hf7/example5_ORU^R03.hl7",
				Output:        "integration_test_results/example5_ORU^R03.hl7",
				Configuration: "conf.test.json",
			},
		}
		for _, testCase := range cases {
			initial, err := os.ReadFile(testCase.Input)
			if err != nil {
				t.Fatalf("message failed to read : %s", err.Error())
			}

			req, err := http.NewRequest("POST", "/not_important", strings.NewReader(string(initial)))
			if err != nil {
				t.Fatalf("Error on creating request: %s", err.Error())
			}

			//record the response
			rr := httptest.NewRecorder()

			//Configure everything
			conf, err := configuration.ParseFile("conf.test.json")
			if err != nil {
				t.Fatalf("Configuration error: %s", err.Error())
			}

			filter := middleware.IngoingFilter{
				ValidConfig:          true,
				HeaderRules:          make(map[string]middleware.FilteringRule),
				URLParameterRules:    make(map[string]middleware.FilteringRule),
				ForbiddenHL7Segments: make(map[string]string),
			}
			err = filter.UpdateWithConfig(conf)
			if err != nil {
				t.Fatalf("Configuration error: %s", err.Error())
			}

			mask := masker.Masker{
				Rules: make(map[string]masker.MaskRule),
			}
			err = mask.UpdateWithConfig(conf)
			if err != nil {
				t.Fatalf("Configuration error: %s", err.Error())
			}

			middlewares := []middleware.Middleware{
				filter.ApplyHeaderRules,
				filter.ApplyParameterRules,
				filter.RejectWithoutValidConfig,
				middleware.LogRequests,
			}

			chainer := middleware.Chainer{
				Middlewares: middlewares,
			}

			local := "http://localhost:4500"

			gHandler := handler.GateHandler{
				Client: &http.Client{},
				Filter: &filter,
				Mask:   &mask,
				Gated:  &local,
			}

			//Now we create a repeater server
			s := &http.Server{
				Addr:    ":4500",
				Handler: handler.RepeatHandler{},
			}
			go func() {
				s.ListenAndServe()
			}()

			chainer.Chain(gHandler).ServeHTTP(rr, req)
			// Check the status code is what we expect.
			if status := rr.Code; status != http.StatusOK {
				t.Fatalf("handler returned wrong status code: got %v want %v",
					status, http.StatusOK)
			}

			expected, err := os.ReadFile(testCase.Output)
			if err != nil {
				t.Fatalf("message failed to read : %s", err.Error())
			}

			if rr.Body.String() != string(expected) {
				// We write in tmp the output to make debugging easier
				os.WriteFile("tmp/old", []byte(expected), 0666)
				os.WriteFile("tmp/new", []byte(rr.Body.String()), 0666)
				t.Fatalf("Error on case : %s", testCase.Input)
			}
			s.Shutdown(context.Background())
		}
	})
}
