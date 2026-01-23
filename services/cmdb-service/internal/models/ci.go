package models

import (
	"database/sql/driver"
	"encoding/json"
	"time"

	"gorm.io/gorm"
)

// JSONB 自定义类型用于存储动态属性
type JSONB map[string]interface{}

func (j JSONB) Value() (driver.Value, error) {
	return json.Marshal(j)
}

func (j *JSONB) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(bytes, j)
}

// CIType CI类型定义
type CIType struct {
	ID          uint           `gorm:"primarykey" json:"id"`
	Name        string         `gorm:"size:50;uniqueIndex;not null" json:"name"`
	DisplayName string         `gorm:"size:100;not null" json:"display_name"`
	Icon        string         `gorm:"size:50" json:"icon"`
	Description string         `gorm:"size:255" json:"description"`
	IsActive    bool           `gorm:"default:true" json:"is_active"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`

	// 关联
	Attributes []CIAttribute `gorm:"foreignKey:CITypeID" json:"attributes,omitempty"`
	Instances  []CIInstance  `gorm:"foreignKey:CITypeID" json:"-"`
}

// CIAttribute CI属性定义
type CIAttribute struct {
	ID          uint           `gorm:"primarykey" json:"id"`
	CITypeID    uint           `gorm:"not null;index" json:"ci_type_id"`
	Name        string         `gorm:"size:50;not null" json:"name"`
	DisplayName string         `gorm:"size:100;not null" json:"display_name"`
	Type        string         `gorm:"size:20;not null" json:"type"` // string, int, float, bool, date, json
	Options     JSONB          `gorm:"type:jsonb" json:"options"`    // 下拉选项、验证规则等
	IsRequired  bool           `gorm:"default:false" json:"is_required"`
	IsUnique    bool           `gorm:"default:false" json:"is_unique"`
	DefaultValue string        `gorm:"size:255" json:"default_value"`
	SortOrder   int            `gorm:"default:0" json:"sort_order"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

// CIInstance CI实例
type CIInstance struct {
	ID         uint           `gorm:"primarykey" json:"id"`
	CITypeID   uint           `gorm:"not null;index" json:"ci_type_id"`
	Name       string         `gorm:"size:100;not null;index" json:"name"`
	Status     string         `gorm:"size:20;default:'active'" json:"status"` // active, inactive, maintenance, decommissioned
	Attributes JSONB          `gorm:"type:jsonb" json:"attributes"`           // 动态属性值
	Tags       JSONB          `gorm:"type:jsonb" json:"tags"`                 // 标签
	CreatedBy  uint           `gorm:"index" json:"created_by"`
	UpdatedBy  uint           `gorm:"index" json:"updated_by"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`

	// 关联
	CIType   *CIType      `gorm:"foreignKey:CITypeID" json:"ci_type,omitempty"`
	Parents  []CIRelation `gorm:"foreignKey:ChildID" json:"-"`
	Children []CIRelation `gorm:"foreignKey:ParentID" json:"-"`
	History  []CIHistory  `gorm:"foreignKey:CIID" json:"-"`
}

// CIRelation CI关系
type CIRelation struct {
	ID           uint           `gorm:"primarykey" json:"id"`
	ParentID     uint           `gorm:"not null;index" json:"parent_id"`
	ChildID      uint           `gorm:"not null;index" json:"child_id"`
	RelationType string         `gorm:"size:50;not null" json:"relation_type"` // depends_on, runs_on, connects_to, contains
	Description  string         `gorm:"size:255" json:"description"`
	CreatedBy    uint           `gorm:"index" json:"created_by"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`

	// 关联
	Parent *CIInstance `gorm:"foreignKey:ParentID" json:"parent,omitempty"`
	Child  *CIInstance `gorm:"foreignKey:ChildID" json:"child,omitempty"`
}

// CIHistory CI变更历史
type CIHistory struct {
	ID        uint      `gorm:"primarykey" json:"id"`
	CIID      uint      `gorm:"not null;index" json:"ci_id"`
	ChangedBy uint      `gorm:"not null;index" json:"changed_by"`
	Action    string    `gorm:"size:20;not null" json:"action"` // create, update, delete
	FieldName string    `gorm:"size:50" json:"field_name"`
	OldValue  string    `gorm:"type:text" json:"old_value"`
	NewValue  string    `gorm:"type:text" json:"new_value"`
	ChangedAt time.Time `json:"changed_at"`
}

// TableName overrides
func (CIType) TableName() string {
	return "ci_types"
}

func (CIAttribute) TableName() string {
	return "ci_attributes"
}

func (CIInstance) TableName() string {
	return "ci_instances"
}

func (CIRelation) TableName() string {
	return "ci_relations"
}

func (CIHistory) TableName() string {
	return "ci_history"
}
