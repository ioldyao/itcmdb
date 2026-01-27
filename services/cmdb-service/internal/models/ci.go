package models

import (
	"context"
	"database/sql/driver"
	"encoding/json"
	"time"

	"gorm.io/gorm"
)

// Context keys for passing data through hooks
const (
	UserIDKey   contextKey = "user_id"
	HistoryKey  contextKey = "history_records"
	SkipHistory contextKey = "skip_history"
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

// GORM Hooks for CIInstance

// BeforeUpdate hook - 检测变更并准备历史记录
func (ci *CIInstance) BeforeUpdate(tx *gorm.DB) error {
	// 检查是否跳过历史记录
	if _, skip := tx.InstanceGet("skip_history"); skip {
		return nil
	}

	ctx := tx.Statement.Context
	if ctx == nil {
		return nil
	}

	// 获取用户ID
	userID, ok := ctx.Value(UserIDKey).(uint)
	if !ok {
		return nil
	}

	// 获取旧数据（从数据库中查询的原始记录）
	var oldCI CIInstance
	if err := tx.First(&oldCI, ci.ID).Error; err != nil {
		return nil
	}

	histories := []CIHistory{}

	// 检查Name字段变化
	if tx.Statement.Changed("Name") {
		histories = append(histories, CIHistory{
			CIID:      ci.ID,
			ChangedBy: userID,
			Action:    "update",
			FieldName: "name",
			OldValue:  oldCI.Name,
			NewValue:  ci.Name,
			ChangedAt: time.Now(),
		})
	}

	// 检查Status字段变化
	if tx.Statement.Changed("Status") {
		histories = append(histories, CIHistory{
			CIID:      ci.ID,
			ChangedBy: userID,
			Action:    "update",
			FieldName: "status",
			OldValue:  oldCI.Status,
			NewValue:  ci.Status,
			ChangedAt: time.Now(),
		})
	}

	// 检查Attributes字段变化（需要深度对比）
	if tx.Statement.Changed("Attributes") {
		oldAttrsJSON, _ := json.Marshal(oldCI.Attributes)
		newAttrsJSON, _ := json.Marshal(ci.Attributes)

		// 只有在JSON字符串不同时才进行深度对比
		if string(oldAttrsJSON) != string(newAttrsJSON) {
			changedFields := getChangedAttributesFiltered(oldCI.Attributes, ci.Attributes)
			if len(changedFields) > 0 {
				histories = append(histories, CIHistory{
					CIID:      ci.ID,
					ChangedBy: userID,
					Action:    "update",
					FieldName: "attributes",
					OldValue:  string(oldAttrsJSON),
					NewValue:  string(newAttrsJSON),
					ChangedAt: time.Now(),
				})
			} else {
				// 没有实际变更（都是被忽略的字段），取消更新
				tx.Statement.Omit("Attributes")
			}
		}
	}

	// 检查Tags字段变化
	if tx.Statement.Changed("Tags") {
		oldTagsJSON, _ := json.Marshal(oldCI.Tags)
		newTagsJSON, _ := json.Marshal(ci.Tags)

		if string(oldTagsJSON) != string(newTagsJSON) {
			histories = append(histories, CIHistory{
				CIID:      ci.ID,
				ChangedBy: userID,
				Action:    "update",
				FieldName: "tags",
				OldValue:  string(oldTagsJSON),
				NewValue:  string(newTagsJSON),
				ChangedAt: time.Now(),
			})
		} else {
			// 没有实际变更，取消更新
			tx.Statement.Omit("Tags")
		}
	}

	// 将历史记录存储到context中，在AfterUpdate中使用
	if len(histories) > 0 {
		tx.InstanceSet("history_records", histories)
	}

	return nil
}

// AfterUpdate hook - 保存历史记录
func (ci *CIInstance) AfterUpdate(tx *gorm.DB) error {
	// 检查是否跳过历史记录
	if _, skip := tx.InstanceGet("skip_history"); skip {
		return nil
	}

	// 从context中获取历史记录
	historyInterface, ok := tx.InstanceGet("history_records")
	if !ok {
		return nil
	}

	histories, ok := historyInterface.([]CIHistory)
	if !ok || len(histories) == 0 {
		return nil
	}

	// 批量创建历史记录
	if err := tx.Create(&histories).Error; err != nil {
		// 记录失败不应该影响主流程，只记录日志
		// TODO: 添加日志记录
		return nil
	}

	return nil
}

// BeforeCreate hook - 记录创建历史
func (ci *CIInstance) BeforeCreate(tx *gorm.DB) error {
	// 检查是否跳过历史记录
	if _, skip := tx.InstanceGet(SkipHistory.String()); skip {
		return nil
	}

	ctx := tx.Statement.Context
	if ctx == nil {
		return nil
	}

	// 获取用户ID
	userID, ok := ctx.Value(UserIDKey).(uint)
	if !ok {
		return nil
	}

	// 准备创建历史记录
	histories := []CIHistory{
		{
			CIID:      ci.ID, // 注意：此时ID可能还未分配，需要在AfterCreate中更新
			ChangedBy: userID,
			Action:    "create",
			FieldName: "",
			OldValue:  "",
			NewValue:  "",
			ChangedAt: time.Now(),
		},
	}

	// 将历史记录存储到context中
	tx.InstanceSet("history_records", histories)

	return nil
}

// AfterCreate hook - 保存创建历史（需要更新CIID）
func (ci *CIInstance) AfterCreate(tx *gorm.DB) error {
	// 检查是否跳过历史记录
	if _, skip := tx.InstanceGet("skip_history"); skip {
		return nil
	}

	// 从context中获取历史记录
	historyInterface, ok := tx.InstanceGet("history_records")
	if !ok {
		return nil
	}

	histories, ok := historyInterface.([]CIHistory)
	if !ok || len(histories) == 0 {
		return nil
	}

	// 更新CIID（此时已经分配了ID）
	for i := range histories {
		histories[i].CIID = ci.ID
	}

	// 创建历史记录
	if err := tx.Create(&histories).Error; err != nil {
		// 记录失败不应该影响主流程
		return nil
	}

	return nil
}

// BeforeDelete hook - 记录删除历史
func (ci *CIInstance) BeforeDelete(tx *gorm.DB) error {
	// 检查是否跳过历史记录
	if _, skip := tx.InstanceGet(SkipHistory.String()); skip {
		return nil
	}

	ctx := tx.Statement.Context
	if ctx == nil {
		return nil
	}

	// 获取用户ID
	userID, ok := ctx.Value(UserIDKey).(uint)
	if !ok {
		return nil
	}

	// 准备删除历史记录
	histories := []CIHistory{
		{
			CIID:      ci.ID,
			ChangedBy: userID,
			Action:    "delete",
			FieldName: "",
			OldValue:  "",
			NewValue:  "",
			ChangedAt: time.Now(),
		},
	}

	// 将历史记录存储到context中
	tx.InstanceSet("history_records", histories)

	return nil
}

// AfterDelete hook - 保存删除历史
func (ci *CIInstance) AfterDelete(tx *gorm.DB) error {
	// 检查是否跳过历史记录
	if _, skip := tx.InstanceGet("skip_history"); skip {
		return nil
	}

	// 从context中获取历史记录
	historyInterface, ok := tx.InstanceGet("history_records")
	if !ok {
		return nil
	}

	histories, ok := historyInterface.([]CIHistory)
	if !ok || len(histories) == 0 {
		return nil
	}

	// 创建历史记录
	if err := tx.Create(&histories).Error; err != nil {
		// 记录失败不应该影响主流程
		return nil
	}

	return nil
}

// getChangedAttributesFiltered 比较两个attributes并返回变化的字段（过滤忽略字段）
func getChangedAttributesFiltered(oldAttrs, newAttrs JSONB) []string {
	changedFields := []string{}

	// 忽略的字段（不记录为变更）
	ignoredFields := map[string]bool{
		"last_hardware_report": true, // 最后上报时间，每次上报都会更新
	}

	// 实时变化的子字段（比较时忽略）
	ignoreSubFields := map[string][]string{
		"optical_modules_info": {"temperature"}, // 光模块温度是实时数据
	}

	// 检查新 attrs 中有哪些字段与旧 attrs 不同
	for key, newValue := range newAttrs {
		// 跳过忽略的字段
		if ignoredFields[key] {
			continue
		}

		oldValue, exists := oldAttrs[key]
		if !exists {
			// 新增的字段
			changedFields = append(changedFields, key+"(新增)")
			continue
		}

		// 比较值是否相同
		oldJSON, _ := json.Marshal(oldValue)
		newJSON, _ := json.Marshal(newValue)

		// 如果直接比较不同，检查是否因为有实时数据字段导致
		if string(oldJSON) != string(newJSON) {
			// 如果该字段有需要忽略的子字段（如 optical_modules_info 的 temperature）
			if subFieldsToIgnore, ok := ignoreSubFields[key]; ok {
				// 尝试排除子字段后再比较
				oldValueCopy := copyValue(oldValue)
				newValueCopy := copyValue(newValue)

				// 移除被忽略的子字段
				if removeSubFields(oldValueCopy, subFieldsToIgnore) || removeSubFields(newValueCopy, subFieldsToIgnore) {
					oldJSONFiltered, _ := json.Marshal(oldValueCopy)
					newJSONFiltered, _ := json.Marshal(newValueCopy)

					if string(oldJSONFiltered) != string(newJSONFiltered) {
						changedFields = append(changedFields, key)
					}
				}
			} else {
				// 没有需要忽略的子字段，直接标记为变化
				changedFields = append(changedFields, key)
			}
		}
	}

	// 检查被删除的字段
	for key := range oldAttrs {
		if _, exists := newAttrs[key]; !exists && !ignoredFields[key] {
			changedFields = append(changedFields, key+"(删除)")
		}
	}

	return changedFields
}

// copyValue 深度复制一个值
func copyValue(value interface{}) interface{} {
	if value == nil {
		return nil
	}

	data, err := json.Marshal(value)
	if err != nil {
		return value
	}

	var result interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		return value
	}

	return result
}

// removeSubFields 从对象或数组中移除指定的子字段
func removeSubFields(value interface{}, fieldsToRemove []string) bool {
	if value == nil {
		return false
	}

	removed := false

	switch v := value.(type) {
	case map[string]interface{}:
		for _, field := range fieldsToRemove {
			if _, ok := v[field]; ok {
				delete(v, field)
				removed = true
			}
		}
	case []interface{}:
		for _, item := range v {
			if itemMap, ok := item.(map[string]interface{}); ok {
				for _, field := range fieldsToRemove {
					if _, ok := itemMap[field]; ok {
						delete(itemMap, field)
						removed = true
					}
				}
			}
		}
	}

	return removed
}

// AutoMigrate 自动迁移表结构
func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&CIType{},
		&CIAttribute{},
		&CIInstance{},
		&CIRelation{},
		&CIHistory{},
		&SystemConfig{},
	)
}
