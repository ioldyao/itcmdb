"""
飞书发送组件
推送告警到飞书
"""

import httpx
from typing import Dict, Any
from pipeline.component.framework.component import Component


class FeishuSenderComponent(Component):
    """飞书发送组件"""

    name = "飞书发送组件"
    code = "sender_feishu"
    form_schema = {
        "type": "object",
        "properties": {
            "webhook_url": {
                "type": "string",
                "title": "Webhook URL"
            }
        }
    }

    def execute(self, data) -> Dict[str, Any]:
        """
        执行发送逻辑

        Inputs:
            - alert: 告警数据（已转换为飞书格式）
            - webhook_url: Webhook URL
        """
        alert = data.get_one_of_inputs("alert", {})
        webhook_url = data.get_one_of_inputs("webhook_url")

        if not webhook_url:
            return self.Outputs(
                success=False,
                message="Feishu webhook URL not provided"
            )

        try:
            # 构建消息
            message = self.build_message(alert)

            # 发送消息
            response = self.send_message(webhook_url, message)

            return self.Outputs(
                success=True,
                message="Alert sent to Feishu successfully",
                response=response
            )

        except Exception as e:
            return self.Outputs(
                success=False,
                message=f"Failed to send alert: {str(e)}",
                error=str(e)
            )

    def build_message(self, alert: Dict[str, Any]) -> Dict[str, Any]:
        """构建飞书消息"""
        # 如果 alert 已经是飞书格式，直接使用
        if "msg_type" in alert:
            message = alert
        else:
            # 否则转换为文本消息
            message = {
                "msg_type": "text",
                "content": {
                    "text": str(alert)
                }
            }

        return message

    def send_message(self, url: str, message: Dict[str, Any]) -> Dict[str, Any]:
        """发送消息"""
        with httpx.Client(timeout=30) as client:
            response = client.post(url, json=message)
            response.raise_for_status()
            return response.json()

    def outputs(self) -> list:
        return [
            {
                "name": "success",
                "type": "boolean",
                "description": "是否成功"
            },
            {
                "name": "message",
                "type": "string",
                "description": "消息"
            },
            {
                "name": "response",
                "type": "object",
                "description": "响应数据"
            },
            {
                "name": "error",
                "type": "string",
                "description": "错误信息"
            }
        ]
