package model

import (
	"context"
	"time"
)

type Config struct {
	Output            string
	CommandTimeout    time.Duration
	IncludeSystemApps bool
}

type Section struct {
	Order   int
	Title   string
	Summary string
	Body    string
	Status  string
}

type Collector struct {
	Order int
	Title string
	Run   func(context.Context, Config) Section
}
