package constants

// Resource 定义系统中所有有效的资源类型
type Resource struct {
	Name        string
	Description string
}

// Action 定义系统中所有有效的操作类型
type Action struct {
	Name        string
	Description string
}

// ValidResources 定义所有有效的资源类型
var ValidResources = []Resource{
	{Name: "user", Description: "用户管理"},
	{Name: "role", Description: "角色管理"},
	{Name: "permission", Description: "权限管理"},
	{Name: "config", Description: "配置管理"},
	{Name: "ci", Description: "配置项管理"},
	{Name: "tag", Description: "标签管理"},
	{Name: "ticket", Description: "工单管理"},
	{Name: "alert", Description: "告警管理"},
	{Name: "audit", Description: "审计日志"},
}

// ValidActions 定义所有有效的操作类型
var ValidActions = []Action{
	{Name: "create", Description: "创建"},
	{Name: "read", Description: "读取"},
	{Name: "update", Description: "更新"},
	{Name: "delete", Description: "删除"},
	{Name: "view", Description: "查看"},
	{Name: "manage", Description: "管理"},
}

// IsValidResource 检查资源是否有效
func IsValidResource(resource string) bool {
	// 特殊权限：超级管理员
	if resource == "*" {
		return true
	}

	for _, r := range ValidResources {
		if r.Name == resource {
			return true
		}
	}
	return false
}

// IsValidAction 检查操作是否有效
func IsValidAction(action string) bool {
	// 特殊权限：超级管理员
	if action == "*" {
		return true
	}

	for _, a := range ValidActions {
		if a.Name == action {
			return true
		}
	}
	return false
}

// GetResourceDescription 获取资源的描述
func GetResourceDescription(resource string) string {
	for _, r := range ValidResources {
		if r.Name == resource {
			return r.Description
		}
	}
	return ""
}

// GetActionDescription 获取操作的描述
func GetActionDescription(action string) string {
	for _, a := range ValidActions {
		if a.Name == action {
			return a.Description
		}
	}
	return ""
}
