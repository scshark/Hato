package crawler

import (
	"github.com/rocboss/paopao-ce/internal/model"
	"gorm.io/gorm"
	"gorm.io/plugin/dbresolver"
)

//type Tabler interface {
//	TableName() string
//}
type TweetUser struct {
	*model.Model
	TweetUserId      string `json:"tweet_user_id"`
	Name             string `json:"name"`
	ScreenName       string `json:"screen_name"`
	Location         string `json:"location"`
	Description      string `json:"description"`
	Urls             string `json:"urls"`
	DescriptionUrls  string `json:"description_urls"`
	ProfileImageUrl  string `json:"profile_image_url"`
	ProfileBannerUrl string `json:"profile_banner_url"`
	TweetCreatedAt   int64  `json:"tweet_created_at"`
	TwitterLoadTime  int64  `json:"twitter_load_time"`
	LoadOlderTime    int64  `json:"load_older_time"`
	LoadType         int64  `json:"load_type"`
	HatoUpdatedAt    int64  `json:"hato_updated_at"`
	FollowersCount   int64  `json:"followers_count"`
	FriendsCount     int64  `json:"friends_count"`
	IsSyncImage      int64  `json:"is_sync_image"`
}

func (t *TweetUser) TableName() string {
	return "ht_twitter_user"
}

func (t *TweetUser) Get(db *gorm.DB) (*TweetUser, error) {
	var user TweetUser
	if t.Model != nil && t.Model.ID > 0 {
		db = db.Where("id= ? AND is_del = ?", t.Model.ID, 0)
	} else if t.ScreenName != "" {
		db = db.Where("screen_name = ? AND is_del = ?", t.ScreenName, 0)
	} else if t.Name != "" {
		db = db.Where("name = ? AND is_del = ?", t.Name, 0)
	}

	err := db.First(&user).Error
	if err != nil {
		return &user, err
	}

	return &user, nil
}

func (t *TweetUser) List(db *gorm.DB, conditions *model.ConditionsT, offset, limit int) ([]*TweetUser, error) {

	var twUser []*TweetUser
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

	if err = db.Where("is_del = ?", 0).Find(&twUser).Error; err != nil {
		return nil, err
	}

	return twUser, nil

}

func (t *TweetUser) Update(db *gorm.DB) error {
	db = db.Clauses(dbresolver.Use("secondary"))
	return db.Model(&TweetUser{}).Where("id = ? AND is_del = ?", t.Model.ID, 0).Save(t).Error
}

func (t *TweetUser) UpdatesTweetUserByIds(ids []int64, up map[string]interface{}, db *gorm.DB) error {

	db = db.Clauses(dbresolver.Use("secondary"))
	return db.Model(&TweetUser{}).Where("id in ? ", ids).Updates(up).Error
}
