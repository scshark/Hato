package model

import (
	"gorm.io/gorm"
)

const (
	UserStatusNormal int = iota + 1
	UserStatusClosed
)

type User struct {
	*Model
	Nickname          string `json:"nickname"`
	Username          string `json:"username"`
	Phone             string `json:"phone"`
	Password          string `json:"password"`
	Salt              string `json:"salt"`
	Status            int    `json:"status"`
	Avatar            string `json:"avatar"`
	BannerUrl         string `json:"banner_url"`
	Balance           int64  `json:"balance"`
	IsAdmin           bool   `json:"is_admin"`
	LoginIp           string `json:"login_ip"`
	Location          string `json:"location"`
	Description       string `json:"description"`
	Urls              string `json:"urls"`
	ProfileBannerUrl  string `json:"profile_banner_url"`
	ProfileImgUrl     string `json:"profile_img_url"`
	FollowersCount    int64  `json:"followers_count"`
	FriendsCount      int64  `json:"friends_count"`
	DescriptionUrls   string `json:"description_urls"`
	PostUpdatedAt     int64  `json:"post_updated_at"`
	IsCrawlerUser     int64  `json:"is_crawler_user"`
	IsCrawlerPlatform int64  `json:"is_crawler_platform"`
	HtTwitterUserId   int64  `json:"ht_twitter_user_id"`
	IsSyncImage       int64  `json:"is_sync_image"`
}

type UserFormated struct {
	ID       int64  `json:"id"`
	Nickname string `json:"nickname"`
	Username string `json:"username"`
	Status   int    `json:"status"`
	Avatar   string `json:"avatar"`
	IsAdmin  bool   `json:"is_admin"`
}

func (u *User) Format() *UserFormated {
	if u.Model != nil {
		return &UserFormated{
			ID:       u.ID,
			Nickname: u.Nickname,
			Username: u.Username,
			Status:   u.Status,
			Avatar:   u.Avatar,
			IsAdmin:  u.IsAdmin,
		}
	}

	return nil
}

func (u *User) Get(db *gorm.DB) (*User, error) {
	var user User
	if u.Model != nil && u.Model.ID > 0 {
		db = db.Where("id= ? AND is_del = ?", u.Model.ID, 0)
	} else if u.Phone != "" {
		db = db.Where("phone = ? AND is_del = ?", u.Phone, 0)
	} else {
		db = db.Where("username = ? AND is_del = ?", u.Username, 0)
	}

	err := db.First(&user).Error
	if err != nil {
		return &user, err
	}

	return &user, nil
}

func (u *User) List(db *gorm.DB, conditions *ConditionsT, offset, limit int) ([]*User, error) {
	var users []*User
	var err error
	if offset >= 0 && limit > 0 {
		db = db.Offset(offset).Limit(limit)
	}
	for k, v := range *conditions {
		if k == "ORDER" {
			db = db.Order(v)
		} else {
			db = db.Where(k, v)
		}
	}

	if err = db.Where("is_del = ?", 0).Find(&users).Error; err != nil {
		return nil, err
	}

	return users, nil
}

func (u *User) Create(db *gorm.DB) (*User, error) {
	err := db.Create(&u).Error

	return u, err
}

func (u *User) Update(db *gorm.DB) error {
	return db.Model(&User{}).Where("id = ? AND is_del = ?", u.Model.ID, 0).Save(u).Error
}

func (u *User) UpdatesUserByIds(ids []int64, up map[string]interface{}, db *gorm.DB) error {

	return db.Model(&User{}).Where("id in ? ", ids).Updates(up).Error
}
