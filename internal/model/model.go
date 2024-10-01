package model

import (
	"encoding/json"
	"time"
)

type Song struct {
	ID          int       `json:"id,omitempty"  example:"1"`
	Group       string    `json:"group" validate:"required" example:"Muse"`
	Song        string    `json:"song" validate:"required" example:"Supermassive Black Hole"`
	ReleaseDate time.Time `json:"releaseDate,omitempty" example:"2006-07-16"`
	Text        string    `json:"text,omitempty" example:"Ooh baby, don't you know I suffer..."`
	Link        string    `json:"link,omitempty" example:"https://www.youtube.com/watch?v=Xsp3_a-PMTw"`
}

func (s *Song) UnmarshalJSON(data []byte) error {
	type Alias Song
	aux := &struct {
		ReleaseDate string `json:"releaseDate,omitempty"`
		*Alias
	}{
		Alias: (*Alias)(s),
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	if aux.ReleaseDate != "" {
		t, err := time.Parse("2006-01-02", aux.ReleaseDate)
		if err != nil {
			return err
		}
		s.ReleaseDate = t
	}
	return nil
}
