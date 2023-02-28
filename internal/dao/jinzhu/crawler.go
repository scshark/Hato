/**
 * @Author: scshark
 * @Description:
 * @File:  crawler
 * @Date: 12/22/22 2:06 PM
 */
package jinzhu

import (
	"fmt"
	"time"

	"github.com/tidwall/gjson"

	"github.com/sirupsen/logrus"

	"github.com/scshark/Hato/internal/core"
	"github.com/scshark/Hato/internal/model"
	"github.com/scshark/Hato/internal/model/crawler"
	"gorm.io/gorm"
)

var (
	_ core.CrawlerService         = (*crawlerServant)(nil)
	_ core.CrawlerTweetService    = (*crawlerTweetServant)(nil)
	_ core.CrawlerPlatformService = (*crawlerPlatformServant)(nil)
)

type crawlerServant struct {
	db *gorm.DB
}
type crawlerTweetServant struct {
	db *gorm.DB
}
type crawlerPlatformServant struct {
	db *gorm.DB
}

func newCrawlerTweetService(db *gorm.DB) core.CrawlerTweetService {
	return &crawlerTweetServant{
		db: db,
	}
}
func newCrawlerService(db *gorm.DB) core.CrawlerService {
	return &crawlerServant{
		db: db,
	}
}
func newCrawlerPlatformService(db *gorm.DB) core.CrawlerPlatformService {
	return &crawlerPlatformServant{
		db: db,
	}
}
func (c *crawlerServant) GetJinseCrawlerData(conditions *model.ConditionsT, offset, limit int) ([]*crawler.Jinse, error) {

	data, err := (&crawler.Jinse{}).List(c.db, conditions, offset, limit)
	return data, err
}

func (c *crawlerServant) UpdateJinseCrawlerData(j *crawler.Jinse) error {
	if err := j.Update(c.db); err != nil {
		fmt.Printf("update error %s", err)
		return err
	}
	return nil
}

func (c *crawlerServant) GetTweetUserByID(id int64) (*crawler.TweetUser, error) {

	user := &crawler.TweetUser{
		Model: &model.Model{
			ID: id,
		},
	}
	return user.Get(c.db)
}

func (c *crawlerServant) GetTweetUserList(conditions *model.ConditionsT, offset, limit int) ([]*crawler.TweetUser, error) {

	data, err := (&crawler.TweetUser{}).List(c.db, conditions, offset, limit)
	return data, err
}
func (c *crawlerServant) UpdateTweetUser(twUser *crawler.TweetUser) error {

	err := twUser.Update(c.db)
	return err
}
func (c *crawlerServant) UpdateTweetUserHatoUpdatedAt(twUserIds []int64) error {
	err := (&crawler.TweetUser{}).UpdatesTweetUserByIds(twUserIds, map[string]interface{}{
		"hato_updated_at":  time.Now().Unix(),
		"need_hato_update": 0,
		"is_sync_image":    1,
	}, c.db)
	return err
}
func (c *crawlerServant) UpdateTweetUserSyncImage(twUserIds []int64) error {
	err := (&crawler.TweetUser{}).UpdatesTweetUserByIds(twUserIds, map[string]interface{}{
		"is_sync_image": 2,
	}, c.db)
	return err
}

func (c *crawlerTweetServant) GetTweetList(conditions *model.ConditionsT, offset, limit int) ([]*crawler.Tweet, error) {

	data, err := (&crawler.Tweet{}).List(c.db, conditions, offset, limit)
	return data, err
}
func (c *crawlerTweetServant) UpdateTweetSyncStatus(twIds []int64) error {
	err := (&crawler.Tweet{}).UpdatesTweetByIds(twIds, map[string]interface{}{
		"is_tweet": 1,
	}, c.db)
	return err
}

func (p *crawlerPlatformServant) GetLivesList(platformType string, conditions *model.ConditionsT, offset, limit int) ([]crawler.PlatformLives, error) {

	var listData = make([]crawler.PlatformLives, 0)
	var err error
	switch platformType {
	case "Jinse":
		list, err := (&crawler.Jinse{}).List(p.db, conditions, offset, limit)
		if err != nil {
			logrus.Errorf("获取金色财经采集数据失败 err %s", err)
			return nil, err
		}

		for _, l := range list {

			var liveItems = make([]crawler.PlatformLivesContent, 0)
			// title
			if l.ContentPrefix != "" {
				liveItems = appendLiveItems(liveItems, l.ContentPrefix, model.CONTENT_TYPE_TITLE)
			}
			//content
			if l.Content != "" {
				liveItems = appendLiveItems(liveItems, l.Content, model.CONTENT_TYPE_TEXT)
			}
			//image
			if l.Images != "" && l.Images != "[]" {
				// parse json
				imageUrl := gjson.Parse(l.Images)
				imageUrl.ForEach(func(key, value gjson.Result) bool {

					if url := value.Get("url").String(); url != "" {
						liveItems = appendLiveItems(liveItems, url, model.CONTENT_TYPE_IMAGE)
					}
					return true
				})
			}

			if l.Link != "" {
				liveItems = appendLiveItems(liveItems, l.Link, model.CONTENT_TYPE_LINK)
			}
			if len(liveItems) > 0 {

				listData = append(listData, crawler.PlatformLives{
					LiveId:    l.Model.ID,
					LiveItems: liveItems,
					CreatedOn: l.LiveCreatedAt,
				})
			}

		}
	case "WallStreet":

		list, err := (&crawler.WallStreet{}).List(p.db, conditions, offset, limit)
		if err != nil {
			logrus.Errorf("获取华尔街见闻采集数据失败 err %s", err)
			return nil, err
		}

		for _, l := range list {

			var liveItems = make([]crawler.PlatformLivesContent, 0)
			// title
			if l.Title != "" {
				liveItems = appendLiveItems(liveItems, l.Title, model.CONTENT_TYPE_TITLE)
			}
			//content
			if l.ContentText != "" {
				liveItems = appendLiveItems(liveItems, l.ContentText, model.CONTENT_TYPE_TEXT)
			}
			//image
			if l.Images != "" && l.Images != "[]" {
				// parse json
				imageUrl := gjson.Parse(l.Images)
				imageUrl.ForEach(func(key, value gjson.Result) bool {

					if url := value.Get("uri").String(); url != "" {
						liveItems = appendLiveItems(liveItems, url, model.CONTENT_TYPE_IMAGE)
					}
					return true
				})
			}

			if l.Uri != "" {
				liveItems = appendLiveItems(liveItems, l.Uri, model.CONTENT_TYPE_LINK)
			}
			if len(liveItems) > 0 {

				listData = append(listData, crawler.PlatformLives{
					LiveId:    l.Model.ID,
					LiveItems: liveItems,
					CreatedOn: l.DisplayTime,
				})
			}

		}
	case "XuanGuBao":
		list, err := (&crawler.Xgb{}).List(p.db, conditions, offset, limit)
		if err != nil {
			logrus.Errorf("获取选股宝采集数据失败 err %s", err)
			return nil, err
		}

		for _, l := range list {

			var liveItems = make([]crawler.PlatformLivesContent, 0)
			// title
			if l.Title != "" {
				liveItems = appendLiveItems(liveItems, l.Title, model.CONTENT_TYPE_TITLE)
			}
			//content
			if l.Summary != "" {
				liveItems = appendLiveItems(liveItems, l.Summary, model.CONTENT_TYPE_TEXT)
			}
			//image
			if l.Image != "" {
				liveItems = appendLiveItems(liveItems, l.Image, model.CONTENT_TYPE_IMAGE)
			}
			// uri
			if l.Uri != "" {
				liveItems = appendLiveItems(liveItems, l.Uri, model.CONTENT_TYPE_LINK)
			}

			var tags = make([]string, 0)
			if l.Tags != "" && l.Tags != "[]" {

				tagsParse := gjson.Parse(l.Tags)

				tagsParse.ForEach(func(key, value gjson.Result) bool {
					tags = append(tags, value.String())
					return true
				})
			}

			if len(liveItems) > 0 {

				listData = append(listData, crawler.PlatformLives{
					LiveId:    l.Model.ID,
					Tags:      tags,
					LiveItems: liveItems,
					CreatedOn: l.LiveCreatedAt,
				})
			}

		}
	case "Dyhjw":
		list, err := (&crawler.Dyhjw{}).List(p.db, conditions, offset, limit)
		if err != nil {
			logrus.Errorf("获取第一黄金网采集数据失败 err %s", err)
			return nil, err
		}

		for _, l := range list {

			var liveItems = make([]crawler.PlatformLivesContent, 0)
			//content
			if l.Content != "" {
				liveItems = appendLiveItems(liveItems, l.Content, model.CONTENT_TYPE_TEXT)
			}

			if len(liveItems) > 0 {
				listData = append(listData, crawler.PlatformLives{
					LiveId:    l.Model.ID,
					LiveItems: liveItems,
					CreatedOn: l.DisplayTime,
				})
			}

		}
	}

	return listData, err
}
func (p *crawlerPlatformServant) UpdateLivesIsTweet(platformType string, liveIds []int64) error {

	var err error
	switch platformType {
	case "Jinse":
		err = (&crawler.Jinse{}).UpdatesByIds(liveIds, map[string]interface{}{
			"is_tweet": 1,
		}, p.db)
	case "WallStreet":
		err = (&crawler.WallStreet{}).UpdatesByIds(liveIds, map[string]interface{}{
			"is_tweet": 1,
		}, p.db)
	case "Dyhjw":
		err = (&crawler.Dyhjw{}).UpdatesByIds(liveIds, map[string]interface{}{
			"is_tweet": 1,
		}, p.db)
	case "XuanGuBao":
		err = (&crawler.Xgb{}).UpdatesByIds(liveIds, map[string]interface{}{
			"is_tweet": 1,
		}, p.db)

	}
	return err
}
func appendLiveItems(items []crawler.PlatformLivesContent, content string, contentType model.PostContentT) []crawler.PlatformLivesContent {

	return append(items, crawler.PlatformLivesContent{
		Content:     content,
		ContentType: contentType,
	})
}
