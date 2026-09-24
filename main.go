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
	// SBB and affiliated operators → SBBP
	"11":     "SBBP", // Schweizerische Bundesbahnen SBB
	"351":    "SBBP", // SBB GmbH (Grenzverkehr)
	"L7____": "SBBP", // SBB GmbH
	"81":     "SBBP", // Aare Seeland mobil (snb)
	"56":     "SBBP", // Aare Seeland mobil (rvo)
	"38":     "SBBP", // Aare Seeland mobil (bti)

	// THURBO
	"65": "THURBO",

	// BLS
	"33": "BLS",

	// RhB
	"72": "RhB",

	// SOB
	"82": "SOB",

	// MBC
	"64": "MBC", // Montreux-Oberland Bernois

	// OeBB
	"81____": "OeBB", // Österreichische Bundesbahnen
	"817000": "OeBB", // NeTS Planung ÖBB

	// TPF
	"53": "TPF", // Transports publics fribourgeois

	// TRN
	"44": "TRN", // Transports Publics Neuchâtelois SA (cmn)
	"73": "TRN", // Transports Publics Neuchâtelois SA (rvt)

	// VDBB
	"06____": "VDBB", // DB Regio AG Baden-Württemberg
	"800693": "VDBB", // DB Regio AG Baden-Württemberg
	"807000": "VDBB", // NeTS Planung DB

	// ZB
	"86": "ZB", // Zentralbahn
}
