package main

import (
	"fmt"
	"log"
	"net/url"
	"os"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v3"
	fiberclient "github.com/gofiber/fiber/v3/client"
	"github.com/joho/godotenv"
)

const formationURL = "https://api.opentransportdata.swiss/formation/v2/formations_full"

// Set at build time with: go build -ldflags "-X main.buildToken=$TOKEN"
var buildToken string

var authToken string

var upstreamClient = fiberclient.New().SetTimeout(30 * time.Second)

type errorResponse struct {
	Error string `json:"error"`
}

func main() {
	if err := godotenv.Load(); err != nil && !os.IsNotExist(err) {
		log.Fatal(fmt.Errorf("load .env: %w", err))
	}

	token := buildToken
	if token == "" {
		token = os.Getenv("TOKEN")
	}
	if token == "" {
		log.Fatal("TOKEN is not set in .env or embedded in the binary")
	}
	authToken = token

	app := fiber.New()

	app.Get("/health", func(c fiber.Ctx) error {
		return c.SendStatus(fiber.StatusNoContent)
	})
	app.Get("/formation", forwardFormation)

	if err := app.Listen(":3001"); err != nil {
		panic(err)
	}
}

func forwardFormation(c fiber.Ctx) error {
	operationDate := c.Query("date")
	evu := c.Query("evu")
	trainNumber := c.Query("trainNumber")

	today := time.Now().UTC().Truncate(24 * time.Hour)
	if operationDate == "" {
		operationDate = today.Format("2006-01-02")
	} else {
		inputDate, err := time.ParseInLocation("2006-01-02", operationDate, time.UTC)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(errorResponse{Error: "date must use the yyyy-mm-dd format"})
		}
		if difference := inputDate.Sub(today); difference > 3*24*time.Hour || difference < -3*24*time.Hour {
			return c.JSON(struct{}{})
		}
	}
	if evu == "" {
		return c.Status(fiber.StatusBadRequest).JSON(errorResponse{Error: "evu is required"})
	}
	if _, err := strconv.Atoi(trainNumber); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(errorResponse{Error: "trainNumber must be a number"})
	}
	apiOperator, ok := GTFSToAPIOperator[evu]
	if !ok {
		return c.JSON(struct{}{})
	}
	requestURL := fmt.Sprintf("%s?evu=%s&operationDate=%s&trainNumber=%s",
		formationURL,
		url.QueryEscape(apiOperator),
		url.QueryEscape(operationDate),
		url.QueryEscape(trainNumber),
	)
	resp, err := upstreamClient.Get(requestURL, fiberclient.Config{
		Header: map[string]string{
			"Authorization": "Bearer " + authToken,
		},
	})
	if err != nil {
		return c.JSON(struct{}{})
	}
	defer resp.Close()

	if resp.StatusCode() < fiber.StatusOK || resp.StatusCode() >= fiber.StatusMultipleChoices {
		return c.JSON(struct{}{})
	}

	c.Status(resp.StatusCode())
	if contentType := resp.Header("Content-Type"); contentType != "" {
		c.Set("Content-Type", contentType)
	}
	return c.Send(resp.Body())
}

var GTFSToAPIOperator = map[string]string{
	// SBB (Swiss Federal Railways)
	"11":     "SBBP",
	"351":    "SBBP",
	"L7____": "SBBP",
	"78":     "SBBP",
	"81":     "SBBP",
	"56":     "SBBP",
	"38":     "SBBP",

	// THURBO
	"65": "THURBO",

	// BLS
	"33": "BLS",

	// RhB (Rhätische Bahn)
	"72": "RhB",

	// SOB (Südostbahn)
	"82": "SOB",

	// MBC (Montreux-Bernois)
	"64":  "MBC",
	"42":  "MBC",
	"131": "MBC",
	"29":  "MBC",

	// OeBB (Österreichische Bundesbahnen)
	"81____": "OeBB",
	"817000": "OeBB",

	// TPF (Transports publics fribourgeois)
	"53": "TPF",

	// TRN (Transports Régionaux)
	"44": "TRN",
	"73": "TRN",

	// VDBB (Verkehrsverbund Baden-Württemberg)
	"06____": "VDBB",
	"800693": "VDBB",
	"807000": "VDBB",

	// ZB (Zentralbahn)
	"86": "ZB",
}
