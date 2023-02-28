package crawler

import (
	"github.com/rocboss/paopao-ce/internal/model"
	"gorm.io/gorm"
	"gorm.io/plugin/dbresolver"
)

type Dyhjw struct {
	*model.Model
	LiveId          string `json:"live_id"`
	Content         string `json:"content"`
	IsTweet         int64  `json:"is_tweet"`
	DisplayTime     int64  `json:"display_time"`
	DisplayDatetime string `json:"display_datetime"`
	Nonce           string `json:"nonce"`
}

func (d *Dyhjw) TableName() string {
	return "ht_dyhjw"
}
func (d *Dyhjw) List(db *gorm.DB, conditions *model.ConditionsT, offset, limit int) ([]*Dyhjw, error) {

	var dh []*Dyhjw
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

	if err = db.Where("is_del = ?", 0).Find(&dh).Error; err != nil {
		return nil, err
	}

	return dh, nil

}
func (d *Dyhjw) UpdatesByIds(ids []int64, up map[string]interface{}, db *gorm.DB) error {

	db = db.Clauses(dbresolver.Use("secondary"))
	return db.Model(&Dyhjw{}).Where("id in ? ", ids).Updates(up).Error
}
