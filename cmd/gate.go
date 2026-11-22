package main

import (
	"fmt"
	"log"
	"net/http"

	"nelia/configuration"
	"nelia/handler"
	"nelia/masker"
	"nelia/middleware"

	"github.com/fsnotify/fsnotify"
)

var gatedServiceHost = "http://localhost:4500"
var configurationPath = "./conf.json"

func main() {
	conf, err := configuration.ParseFile(configurationPath)
	if err != nil {
		log.Fatal(err)
	}

	filter := middleware.IngoingFilter{
		ValidConfig:          true,
		HeaderRules:          make(map[string]middleware.FilteringRule),
		URLParameterRules:    make(map[string]middleware.FilteringRule),
		ForbiddenHL7Segments: make(map[string]string),
	}
	err = filter.UpdateWithConfig(conf)
	if err != nil {
		log.Fatal(err)
	}

	mask := masker.Masker{
		Rules: make(map[string]masker.MaskRule),
	}
	err = mask.UpdateWithConfig(conf)
	if err != nil {
		log.Fatal(err)
	}

	// Create new watcher, so we know when to reload the conf
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	// Start listening for events.
	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				if event.Has(fsnotify.Create) && event.Name == configurationPath {
					filter.Disable()
					conf, err := configuration.ParseFile(configurationPath)
					if err != nil {
						log.Print(fmt.Errorf("Error updating configuration: %w", err))
						break
					}

					err = filter.UpdateWithConfig(conf)
					if err != nil {
						log.Print(fmt.Errorf("Error updating filter configuration: %w", err))
						break
					}

					err = mask.UpdateWithConfig(conf)
					if err != nil {
						log.Print(fmt.Errorf("Error updating maskingconfiguration: %w", err))
						break
					}
					gatedServiceHost = conf.GatedURL
					filter.Enable()
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Println("watcher error:", err)
			}
		}
	}()

	err = watcher.Add("./")
	if err != nil {
		log.Fatal(err)
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

	handler := handler.GateHandler{
		Client: &http.Client{},
		Filter: &filter,
		Mask:   &mask,
		Gated:  &gatedServiceHost,
	}

	log.Println("Application has started. Listening port is 4400")
	err = http.ListenAndServe(":4400", chainer.Chain(handler))
	if err != nil {
		log.Fatal(err)
	}
}
