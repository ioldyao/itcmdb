"""
企业微信发送组件
推送告警到企业微信
"""

import httpx
from typing import Dict, Any
from pipeline.component.framework.component import Component


class WechatSenderComponent(Component):
    """企业微信发送组件"""

    name = "企业微信发送组件"
    code = "sender_wechat"
    form_schema = {
        "type": "object",
        "properties": {
            "webhook_url": {
                "type": "string",
                "title": "Webhook URL"
            },
            "mentioned_list": {
                "type": "array",
                "title": "@用户列表",
                "items": {"type": "string"}
            },
            "mentioned_mobile_list": {
                "type": "array",
                "title": "@手机号列表",
                "items": {"type": "string"}
            }
        }
    }

    def execute(self, data) -> Dict[str, Any]:
        """
        执行发送逻辑

        Inputs:
            - alert: 告警数据（已转换为企业微信格式）
            - webhook_url: Webhook URL
            - mentioned_list: @用户列表
            - mentioned_mobile_list: @手机号列表
        """
        alert = data.get_one_of_inputs("alert", {})
        webhook_url = data.get_one_of_inputs("webhook_url")
        mentioned_list = data.get_one_of_inputs("mentioned_list", [])
        mentioned_mobile_list = data.get_one_of_inputs("mentioned_mobile_list", [])

        if not webhook_url:
            return self.Outputs(
                success=False,
                message="Wechat webhook URL not provided"
            )

        try:
            # 构建消息
            message = self.build_message(alert, mentioned_list, mentioned_mobile_list)

            # 发送消息
            response = self.send_message(webhook_url, message)

            return self.Outputs(
                success=True,
                message="Alert sent to Wechat successfully",
                response=response
            )

        except Exception as e:
            return self.Outputs(
                success=False,
                message=f"Failed to send alert: {str(e)}",
                error=str(e)
            )

    def build_message(self, alert: Dict[str, Any], mentioned_list: list, mentioned_mobile_list: list) -> Dict[str, Any]:
        """构建企业微信消息"""
        # 如果 alert 已经是企业微信格式，直接使用
        if "msgtype" in alert:
            message = alert
            # 添加 @ 信息
            if message["msgtype"] == "text":
                if "text" not in message:
                    message["text"] = {}
                if mentioned_list:
                    message["text"]["mentioned_list"] = mentioned_list
                if mentioned_mobile_list:
                    message["text"]["mentioned_mobile_list"] = mentioned_mobile_list
        else:
            # 否则转换
            message = {
                "msgtype": "text",
                "text": {
                    "content": str(alert),
                    "mentioned_list": mentioned_list,
                    "mentioned_mobile_list": mentioned_mobile_list
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
