package model

import "time"

type Song struct {
	ID          int       `json:"id,omitempty"  example:"1"`
	Group       string    `json:"group" validate:"required" example:"Muse"`
	Song        string    `json:"song" validate:"required" example:"Supermassive Black Hole"`
	ReleaseDate time.Time `json:"releaseDate,omitempty" example:"2006-07-16T00:00:00Z"`
	Text        string    `json:"text,omitempty" example:"Ooh baby, don't you know I suffer..."`
	Link        string    `json:"link,omitempty" example:"https://www.youtube.com/watch?v=Xsp3_a-PMTw"`
}
