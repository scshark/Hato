/**
 * @Author: scshark
 * @Description:
 * @File:  jinse
 * @Date: 12/22/22 2:23 PM
 */
package crawler

import (
	"gorm.io/plugin/dbresolver"

	"github.com/rocboss/paopao-ce/internal/model"
	"gorm.io/gorm"
)

type Jinse struct {
	*model.Model
	TopId         int64  `json:"top_id"`
	BottomId      int64  `json:"bottom_id"`
	LiveId        int64  `json:"live_id"`
	Content       string `json:"content"`
	ContentPrefix string `json:"content_prefix"`
	Images        string `json:"images"`
	LinkName      string `json:"link_name"`
	Link          string `json:"link"`
	LiveCreatedAt int64  `json:"live_created_at"`
	CreatedAtZh   string `json:"created_at_zh" `
	IsTweet       int64  `json:"is_tweet" `
}

func (j *Jinse) TableName() string {
	return "ht_jinse"
}
func (j *Jinse) List(db *gorm.DB, conditions *model.ConditionsT, offset, limit int) ([]*Jinse, error) {

	var js []*Jinse
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

	if err = db.Where("is_del = ?", 0).Find(&js).Error; err != nil {
		return nil, err
	}

	return js, nil

}
func (j *Jinse) Update(db *gorm.DB) error {
	db = db.Clauses(dbresolver.Use("secondary"))
	return db.Model(&Jinse{}).Where("id = ? AND is_del = ?", j.Model.ID, 0).Save(j).Error
}

func (j *Jinse) UpdatesByIds(ids []int64, up map[string]interface{}, db *gorm.DB) error {

	db = db.Clauses(dbresolver.Use("secondary"))
	return db.Model(&Jinse{}).Where("id in ? ", ids).Updates(up).Error
}
