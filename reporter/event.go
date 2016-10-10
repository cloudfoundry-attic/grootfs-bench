package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"
)

var eventsEndpoint = "https://app.datadoghq.com/api/v1/events?api_key=" + os.Getenv("DATADOG_API_KEY") + "&application_key=" + os.Getenv("DATADOG_APPLICATION_KEY")

const eventTag = "grootfs:performance"
const eventTitleTemplate = "grootfs-commit: %s"

func eventCommand() error {
	eventTitle = fmt.Sprintf(eventTitleTemplate, eventTitle)

	alreadyPublished, err := eventAlreadyPublished(eventTitle)
	if err != nil {
		return err
	}

	if alreadyPublished {
		fmt.Println("Already published")
		return nil
	}

	fmt.Println("publishing event")
	return publishEvent(eventTitle, eventMessage)
}

type event struct {
	Title        string   `json:"title"`
	Text         string   `json:"text"`
	Tags         []string `json:"tags"`
	DateHappened int64    `json:"date_happened"`
}

func publishEvent(title, message string) error {
	event := event{
		Title:        title,
		Text:         message,
		Tags:         []string{eventTag},
		DateHappened: time.Now().Unix(),
	}

	buffer := new(bytes.Buffer)
	err := json.NewEncoder(buffer).Encode(event)
	if err != nil {
		return err
	}

	response, err := http.Post(eventsEndpoint, "application/json", buffer)
	if err != nil {
		return fmt.Errorf("Submit event:", err)
	}

	if response.StatusCode != http.StatusAccepted {
		return fmt.Errorf("Submit event returned status code %d", response.StatusCode)
	}

	return nil
}

type eventQueryResponse struct {
	Events []event `json:"events"`
}

func eventAlreadyPublished(title string) (bool, error) {
	end := time.Now()
	start := end.Add(-30 * 24 * time.Hour)

	response, err := http.Get(fmt.Sprintf("%s&start=%d&end=%d", eventsEndpoint, start.Unix(), end.Unix()))
	if err != nil {
		return false, fmt.Errorf("Query for events:", err)
	}

	if response.StatusCode != http.StatusOK {
		var body []byte
		response.Body.Read(body)
		return false, fmt.Errorf("Event query returned status code %d", response.StatusCode)
	}

	var events eventQueryResponse
	err = json.NewDecoder(response.Body).Decode(&events)
	if err != nil {
		return false, err
	}

	for _, event := range events.Events {
		if event.Title == title {
			return true, nil
		}
	}

	return false, nil
}
