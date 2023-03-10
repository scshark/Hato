package model

import (
	"time"

	"github.com/scshark/Hato/pkg/types"
	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"
)

// Model 公共Model
type Model struct {
	ID         int64                 `gorm:"primary_key" json:"id"`
	CreatedOn  int64                 `json:"created_on"`
	ModifiedOn int64                 `json:"modified_on"`
	DeletedOn  int64                 `json:"deleted_on"`
	IsDel      soft_delete.DeletedAt `gorm:"softDelete:flag" json:"is_del"`
}

type ConditionsT map[string]interface{}
type Predicates map[string]types.AnySlice

func (m *Model) BeforeCreate(tx *gorm.DB) (err error) {

	if m != nil && m.CreatedOn > 0 {
		return
	}

	nowTime := time.Now().Unix()
	tx.Statement.SetColumn("created_on", nowTime)
	tx.Statement.SetColumn("modified_on", nowTime)
	return
}

func (m *Model) BeforeUpdate(tx *gorm.DB) (err error) {
	if !tx.Statement.Changed("modified_on") {
		tx.Statement.SetColumn("modified_on", time.Now().Unix())
	}

	return
}
