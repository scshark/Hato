package jinzhu

import (
	"strings"

	"github.com/scshark/Hato/internal/core"
	"github.com/scshark/Hato/internal/model"
	"gorm.io/gorm"
)

var (
	_ core.UserManageService = (*userManageServant)(nil)
)

type userManageServant struct {
	db *gorm.DB
}

func newUserManageService(db *gorm.DB) core.UserManageService {
	return &userManageServant{
		db: db,
	}
}

func (s *userManageServant) GetUserByID(id int64) (*model.User, error) {
	user := &model.User{
		Model: &model.Model{
			ID: id,
		},
	}
	return user.Get(s.db)
}

func (s *userManageServant) GetUserByUsername(username string) (*model.User, error) {
	user := &model.User{
		Username: username,
	}
	return user.Get(s.db)
}

func (s *userManageServant) GetUserByPhone(phone string) (*model.User, error) {
	user := &model.User{
		Phone: phone,
	}
	return user.Get(s.db)
}

func (s *userManageServant) GetUsersByIDs(ids []int64) ([]*model.User, error) {
	user := &model.User{}
	return user.List(s.db, &model.ConditionsT{
		"id IN ?": ids,
	}, 0, 0)
}

func (s *userManageServant) GetUsersByKeyword(keyword string) ([]*model.User, error) {
	user := &model.User{}
	keyword = strings.Trim(keyword, " ") + "%"
	if keyword == "%" {
		return user.List(s.db, &model.ConditionsT{
			"ORDER": "id ASC",
		}, 0, 6)
	} else {
		return user.List(s.db, &model.ConditionsT{
			"username LIKE ?": keyword,
		}, 0, 6)
	}
}

func (s *userManageServant) GetTagsByKeyword(keyword string) ([]*model.Tag, error) {
	tag := &model.Tag{}
	keyword = "%" + strings.Trim(keyword, " ") + "%"
	if keyword == "%%" {
		return tag.List(s.db, &model.ConditionsT{
			"ORDER": "quote_num DESC",
		}, 0, 6)
	} else {
		return tag.List(s.db, &model.ConditionsT{
			"tag LIKE ?": keyword,
			"ORDER":      "quote_num DESC",
		}, 0, 6)
	}
}

func (s *userManageServant) CreateUser(user *model.User) (*model.User, error) {
	return user.Create(s.db)
}

func (s *userManageServant) UpdateUser(user *model.User) error {
	return user.Update(s.db)
}

func (s *userManageServant) IsFriend(userId int64, friendId int64) bool {
	// just true now
	return true
}

func (s *userManageServant) GetUserList(conditions *model.ConditionsT, offset, limit int) ([]*model.User, error) {
	user := &model.User{}

	return user.List(s.db, conditions, offset, limit)
}

func (s *userManageServant) UpdateUserSyncImage(userIds []int64) error {
	err := (&model.User{}).UpdatesUserByIds(userIds, map[string]interface{}{
		"is_sync_image": 2,
	}, s.db)
	return err
}
