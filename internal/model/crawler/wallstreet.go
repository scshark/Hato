package crawler

import (
	"github.com/scshark/Hato/internal/model"
	"gorm.io/gorm"
	"gorm.io/plugin/dbresolver"
)

type WallStreet struct {
	*model.Model
	Title       string `json:"title"`
	Uri         string `json:"uri"`
	DisplayTime int64  `json:"display_time"`
	CoverImages string `json:"cover_images"`
	Content     string `json:"content"`
	ContentText string `json:"content_text"`
	ContentMore string `json:"content_more" `
	Images      string `json:"images" `
	Author      string `json:"author" `
	IsTweet     int64  `json:"is_tweet" `
}

func (w *WallStreet) TableName() string {
	return "ht_wallstreet"
}
func (w *WallStreet) List(db *gorm.DB, conditions *model.ConditionsT, offset, limit int) ([]*WallStreet, error) {

	var ws []*WallStreet
	var err error

	db = db.Clauses(dbresolver.Use("secondary"))

	if offset != 0 {
		db = db.Offset(offset)
	}
	if limit > 0 {
		db = db.Limit(limit)
	}

	for k, v := range *conditions {
		if k == "ORDER" {
			db = db.Order(v)
		} else {
			db = db.Where(k, v)
		}
	}

	if err = db.Where("is_del = ?", 0).Find(&ws).Error; err != nil {
		return nil, err
	}

	return ws, nil

}

func (w *WallStreet) UpdatesByIds(ids []int64, up map[string]interface{}, db *gorm.DB) error {

	db = db.Clauses(dbresolver.Use("secondary"))
	return db.Model(&WallStreet{}).Where("id in ? ", ids).Updates(up).Error
}
