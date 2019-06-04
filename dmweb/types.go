package dmweb

import (
	"net/http"
	"time"
)

// Client represents a DMWeb API client
type Client struct {
	Client    *http.Client
	AccountID string
	Username  string
	Password  string
	DevID     string
	baseURL   string
	userAgent string
}

// Tag represents an EWON tag
type Tag struct {
	ID          int     `json:"id"`
	Name        string  `json:"name"`
	DataType    string  `json:"dataType"`
	Description string  `json:"description"`
	AlarmHint   string  `json:"alarmHint"`
	Value       float64 `json:"value"`
	Quality     string  `json:"quality"`
	EwonTagID   int     `json:"ewonTagId"`
}

type Tags []*Tag

// Ewon defines an ewon object
// timeZone is an optional field
type Ewon struct {
	ID              int       `json:"id"`
	Name            string    `json:"name"`
	LastSynchroDate time.Time `json:"lastSynchroDate"`
	Tags            Tags      `json:"tags"`
	TimeZone        string    `json:"timeZone"`
}

// Ewons represents multiple Ewon
type Ewons []*Ewon

// GetStatusResponse represents a status response
type GetStatusResponse struct {
	HistoryCount int `json:"historyCount"`
	EwonsCount   int `json:"ewonsCount"`
	Ewons        []struct {
		ID               int       `json:"id"`
		Name             string    `json:"name"`
		HistoryCount     int       `json:"historyCount"`
		FirstHistoryDate time.Time `json:"firstHistoryDate"`
		LastHistoryDate  time.Time `json:"lastHistoryDate"`
	} `json:"ewons"`
}

// GetDataResponse represents a successful response
// to the getdata endpoint
type GetDataResponse struct {
	Success           bool `json:"success"`
	MoreDataAvailable bool `json:"moreDataAvailable"`
	Ewons             []struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
		Tags []struct {
			ID          int    `json:"id"`
			Name        string `json:"name"`
			DataType    string `json:"dataType"`
			Description string `json:"description"`
			AlarmHint   string `json:"alarmHint"`
			Value       int    `json:"value"`
			Quality     string `json:"quality"`
			EwonTagID   int    `json:"ewonTagId"`
			History     []struct {
				Date    time.Time `json:"date,omitempty"`
				Quality string    `json:"quality,omitempty"`
				Value   int       `json:"value"`
			} `json:"history"`
		} `json:"tags"`
		LastSynchroDate time.Time `json:"lastSynchroDate"`
		TimeZone        string    `json:"timeZone"`
	} `json:"ewons"`
}

// SyncResponse represents a successful response
// to the syncdata endpoint.
type SyncResponse struct {
	Success           bool   `json:"success"`
	TransactionID     string `json:"transactionId"`
	MoreDataAvailable bool   `json:"moreDataAvailable"`
	Ewons             []struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
		Tags []struct {
			ID          int     `json:"id"`
			Name        string  `json:"name"`
			DataType    string  `json:"dataType"`
			Description string  `json:"description"`
			AlarmHint   string  `json:"alarmHint"`
			Value       float64 `json:"value"`
			Quality     string  `json:"quality"`
			EwonTagID   int     `json:"ewonTagId"`
			History     []struct {
				Date     time.Time `json:"date"`
				DataType string    `json:"dataType"`
				Value    float64   `json:"value"`
			} `json:"history"`
		} `json:"tags"`
		LastSynchroDate time.Time `json:"lastSynchroDate"`
	} `json:"ewons"`
}

type errorResponse struct {
	Success bool   `json:"success"`
	Code    int    `json:"code"`
	Message string `json:"message"`
}
