package crawler

import "github.com/rocboss/paopao-ce/internal/model"

type Platform struct {
	Jinse      Jinse
	WallStreet WallStreet
	Dyhjw      Dyhjw
	Xgb        Xgb
}
type PlatformLives struct {
	LiveId    int64                  `json:"live_id"`
	LiveItems []PlatformLivesContent `json:"live_items"`
	Tags      []string               `json:"tags"`
	CreatedOn int64                  `json:"created_on"`
}
type PlatformLivesContent struct {
	Content     string             `json:"content"`
	ContentType model.PostContentT `json:"content_type"`
}
