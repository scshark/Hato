package crawler

import (
	"github.com/scshark/Hato/internal/model"
	"gorm.io/gorm"
	"gorm.io/plugin/dbresolver"
)

type Xgb struct {
	*model.Model
	Title         string `json:"title"`
	Summary       string `json:"summary"`
	Image         string `json:"image"`
	LiveCreatedAt int64  `json:"live_created_at"`
	SubjIds       string `json:"subj_ids"`
	Uri           string `json:"uri"`
	Tags          string `json:"tags"`
	OriginaUrl    string `json:"origina_url"`
	Source        string `json:"source"`
	IsTweet       int64  `json:"is_tweet" `
}

func (x *Xgb) TableName() string {
	return "ht_xgb"
}
func (x *Xgb) List(db *gorm.DB, conditions *model.ConditionsT, offset, limit int) ([]*Xgb, error) {

	var xgb []*Xgb
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

	if err = db.Where("is_del = ?", 0).Find(&xgb).Error; err != nil {
		return nil, err
	}

	return xgb, nil

}

func (x *Xgb) UpdatesByIds(ids []int64, up map[string]interface{}, db *gorm.DB) error {

	db = db.Clauses(dbresolver.Use("secondary"))
	return db.Model(&Xgb{}).Where("id in ? ", ids).Where("is_del = ? ", 0).Updates(up).Error
}
