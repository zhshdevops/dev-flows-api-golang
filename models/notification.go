package models

import (
	"time"
)

// Notification holds all events.
type Notification struct {
	Events []Event
}

// Event holds the details of a event.
type Event struct {
	ID        string `json:"Id"`
	TimeStamp time.Time `json:"timestamp"`
	Action    string `json:"action"`
	Target    *Target `json:"target"`
	Request   *Request `json:"request"`
	Actor     *Actor `json:"actor"`
	Source    *Source `json:"source"`
}

type Source struct {
	Addr       string `json:"addr"`
	InstanceID string `json:"instanceID"`
}

// Target holds information about the target of a event.
type Target struct {
	MediaType  string `json:"mediaType"`
	Digest     string `json:"digest"`
	Repository string `json:"repository"`
	URL        string `json:"url"`
	Tag        string `json:"tag"`
	Length     int `json:"length"`
	Size       int `json:"size"`
}

// Actor holds information about actor.
type Actor struct {
	Name string `json:"name"`
}

// Request holds information about a request.
type Request struct {
	ID        string `json:"Id"`
	Method    string `json:"method"`
	UserAgent string `json:"useragent"`
	Host      string `json:"host"`
	Addr      string `json:"addr"`
}

type ImageInfo struct {
	Fullname    string
	Projectname string //镜像仓库
	Tag         string
}
