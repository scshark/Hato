/**
 * @Author: scshark
 * @Description:
 * @File:  crawler
 * @Date: 12/22/22 2:11 PM
 */
package core

import (
	"github.com/scshark/Hato/internal/model"
	"github.com/scshark/Hato/internal/model/crawler"
)

type CrawlerService interface {
	GetJinseCrawlerData(conditions *model.ConditionsT, offset, limit int) ([]*crawler.Jinse, error)
	UpdateJinseCrawlerData(*crawler.Jinse) error

	// tweet user
	GetTweetUserByID(id int64) (*crawler.TweetUser, error)
	GetTweetUserList(conditions *model.ConditionsT, offset, limit int) ([]*crawler.TweetUser, error)
	UpdateTweetUser(twUser *crawler.TweetUser) error
	UpdateTweetUserHatoUpdatedAt(twUserIds []int64) error
	UpdateTweetUserSyncImage(twUserIds []int64) error
}

// Tweet
type CrawlerTweetService interface {
	GetTweetList(conditions *model.ConditionsT, offset, limit int) ([]*crawler.Tweet, error)
	UpdateTweetSyncStatus(twIds []int64) error
}

type CrawlerPlatformService interface {
	GetLivesList(platformType string, conditions *model.ConditionsT, offset, limit int) ([]crawler.PlatformLives, error)
	UpdateLivesIsTweet(platformType string, liveIds []int64) error
}
