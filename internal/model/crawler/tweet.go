package crawler

import (
	"github.com/rocboss/paopao-ce/internal/model"
	"gorm.io/gorm"
	"gorm.io/plugin/dbresolver"
)

type Tweet struct {
	*model.Model
	HtTwitterUserId  int64  `json:"ht_twitter_user_id"`
	TwitterUserId    string `json:"twitter_user_id"`
	IdStr            string `json:"id_str"`
	FullText         string `json:"full_text"`
	Hashtags         string `json:"hashtags"`
	UserMentions     string `json:"user_mentions"`
	Urls             string `json:"urls"`
	ExtendedEntities string `json:"extended_entities"`
	InReplyInfo      string `json:"in_reply_info"`
	TwCreatedAt      int64  `json:"tw_created_at"`
	IsTweet          int64  `json:"is_tweet" `
}

func (t *Tweet) TableName() string {
	return "ht_twitter"
}

func (t *Tweet) List(db *gorm.DB, conditions *model.ConditionsT, offset, limit int) ([]*Tweet, error) {

	var tweet []*Tweet
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

	if err = db.Where("is_del = ?", 0).Find(&tweet).Error; err != nil {
		return nil, err
	}

	return tweet, nil

}

func (t *Tweet) UpdatesTweetByIds(ids []int64, up map[string]interface{}, db *gorm.DB) error {

	db = db.Clauses(dbresolver.Use("secondary"))
	return db.Model(&Tweet{}).Where("id in ? ", ids).Updates(up).Error
}
