package rest

import "github.com/scshark/Hato/internal/model"

type IndexTweetsResp struct {
	Tweets []*model.PostFormated
	Total  int64
}
