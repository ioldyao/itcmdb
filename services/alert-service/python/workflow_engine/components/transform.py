"""
告警转换组件
转换告警格式
"""

from typing import Dict, Any
from pipeline.component.framework.component import Component


class AlertTransformComponent(Component):
    """告警转换组件"""

    name = "告警转换组件"
    code = "alert_transform"
    form_schema = {
        "type": "object",
        "properties": {
            "target_format": {
                "type": "string",
                "enum": ["dingtalk", "wechat", "feishu", "email", "custom"],
                "title": "目标格式"
            },
            "template": {
                "type": "string",
                "title": "自定义模板"
            },
            "field_mapping": {
                "type": "object",
                "title": "字段映射"
            }
        }
    }

    def execute(self, data) -> Dict[str, Any]:
        """
        执行转换逻辑

        Inputs:
            - alert: 告警数据
            - target_format: 目标格式
            - template: 自定义模板
            - field_mapping: 字段映射
        """
        alert = data.get_one_of_inputs("alert", {})
        target_format = data.get_one_of_inputs("target_format", "custom")
        template = data.get_one_of_inputs("template")
        field_mapping = data.get_one_of_inputs("field_mapping", {})

        # 应用字段映射
        mapped_alert = self.apply_field_mapping(alert, field_mapping)

        # 转换为目标格式
        if target_format == "dingtalk":
            transformed = self.transform_to_dingtalk(mapped_alert, template)
        elif target_format == "wechat":
            transformed = self.transform_to_wechat(mapped_alert, template)
        elif target_format == "feishu":
            transformed = self.transform_to_feishu(mapped_alert, template)
        elif target_format == "email":
            transformed = self.transform_to_email(mapped_alert, template)
        else:
            transformed = mapped_alert

        return self.Outputs(
            success=True,
            transformed_alert=transformed,
            original_alert=alert
        )

    def apply_field_mapping(self, alert: Dict[str, Any], mapping: Dict[str, str]) -> Dict[str, Any]:
        """应用字段映射"""
        if not mapping:
            return alert

        result = {}
        for target_field, source_field in mapping.items():
            value = self.get_field_value(alert, source_field)
            result[target_field] = value

        return result

    def get_field_value(self, alert: Dict[str, Any], field: str) -> Any:
        """获取字段值（支持嵌套）"""
        keys = field.split(".")
        value = alert

        for key in keys:
            if isinstance(value, dict):
                value = value.get(key)
            else:
                return None

        return value

    def transform_to_dingtalk(self, alert: Dict[str, Any], template: str = None) -> Dict[str, Any]:
        """转换为钉钉格式"""
        if template:
            text = template.format(**alert)
        else:
            text = (
                f"#### {alert.get('title', '告警通知')}\n"
                f"> **状态**: {alert.get('status', 'firing')}\n"
                f"> **级别**: {alert.get('severity', 'warning')}\n"
                f"> **时间**: {alert.get('startsAt', '')}\n"
                f"> **详情**:\n{alert.get('description', '')}\n"
            )

        return {
            "msgtype": "markdown",
            "markdown": {
                "title": alert.get("title", "告警通知"),
                "text": text
            }
        }

    def transform_to_wechat(self, alert: Dict[str, Any], template: str = None) -> Dict[str, Any]:
        """转换为企业微信格式"""
        if template:
            content = template.format(**alert)
        else:
            content = (
                f"### {alert.get('title', '告警通知')}\n"
                f"**状态**: {alert.get('status', 'firing')}\n"
                f"**级别**: {alert.get('severity', 'warning')}\n"
                f"**时间**: {alert.get('startsAt', '')}\n"
                f"**详情**: {alert.get('description', '')}\n"
            )

        return {
            "msgtype": "markdown",
            "markdown": {
                "content": content
            }
        }

    def transform_to_feishu(self, alert: Dict[str, Any], template: str = None) -> Dict[str, Any]:
        """转换为飞书格式"""
        content = [
            {
                "tag": "text",
                "text": f"{alert.get('title', '告警通知')}\n"
            },
            {
                "tag": "text",
                "text": f"状态: {alert.get('status', 'firing')}\n"
            },
            {
                "tag": "text",
                "text": f"级别: {alert.get('severity', 'warning')}\n"
            },
            {
                "tag": "text",
                "text": f"详情: {alert.get('description', '')}\n"
            }
        ]

        return {
            "msg_type": "post",
            "content": {
                "post": {
                    "zh_cn": {
                        "title": alert.get("title", "告警通知"),
                        "content": content
                    }
                }
            }
        }

    def transform_to_email(self, alert: Dict[str, Any], template: str = None) -> Dict[str, Any]:
        """转换为邮件格式"""
        return {
            "subject": alert.get("title", "告警通知"),
            "body": template.format(**alert) if template else str(alert),
            "html": False
        }

    def outputs(self) -> list:
        return [
            {
                "name": "success",
                "type": "boolean",
                "description": "是否成功"
            },
            {
                "name": "transformed_alert",
                "type": "object",
                "description": "转换后的告警"
            },
            {
                "name": "original_alert",
                "type": "object",
                "description": "原始告警"
            }
        ]
