package model

type Song struct {
	ID          int    `json:"id,omitempty"  example:"1"`
	Group       string `json:"group,omitempty" validate:"required" example:"Muse"`
	Song        string `json:"song,omitempty" validate:"required" example:"Supermassive Black Hole"`
	ReleaseDate string `json:"releaseDate,omitempty" example:"2006-07-16"`
	Text        string `json:"text,omitempty" example:"Ooh baby, don't you know I suffer..."`
	Link        string `json:"link,omitempty" example:"https://www.youtube.com/watch?v=Xsp3_a-PMTw"`
}
