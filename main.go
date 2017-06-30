package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/pkg/errors"
)

const version = "1.0"

func main() {
	logger := log.New(os.Stderr, "", log.Lshortfile)
	logger.Printf("simple-autoscaler %s starting", version)

	var rules []Rule
	err := json.Unmarshal([]byte(os.Getenv("AUTOSCALER_RULES")), &rules)
	if err != nil {
		logger.Fatal(errors.Wrap(err, "parse autoscaler rules"))
	}

	go func() {
		http.ListenAndServe(":"+os.Getenv("PORT"), nil)
	}()

	run(Config{
		Logger:            logger,
		ApiUrl:            os.Getenv("CF_API_URL"),
		ApiUsername:       os.Getenv("CF_USERNAME"),
		ApiPassword:       os.Getenv("CF_PASSWORD"),
		SkipSslValidation: os.Getenv("SKIP_SSL_VALIDATION") == "true",
		Rules:             rules,
	})
}
