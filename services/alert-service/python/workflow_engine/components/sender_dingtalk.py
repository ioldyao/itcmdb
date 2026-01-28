"""
钉钉发送组件
推送告警到钉钉
"""

import httpx
import hmac
import hashlib
import base64
import time
from typing import Dict, Any
from pipeline.component.framework.component import Component


class DingtalkSenderComponent(Component):
    """钉钉发送组件"""

    name = "钉钉发送组件"
    code = "sender_dingtalk"
    form_schema = {
        "type": "object",
        "properties": {
            "webhook_url": {
                "type": "string",
                "title": "Webhook URL"
            },
            "secret": {
                "type": "string",
                "title": "签名密钥（可选）"
            },
            "at_mobiles": {
                "type": "array",
                "title": "@手机号列表",
                "items": {"type": "string"}
            },
            "at_user_ids": {
                "type": "array",
                "title": "@用户ID列表",
                "items": {"type": "string"}
            },
            "is_at_all": {
                "type": "boolean",
                "title": "@所有人",
                "default": False
            }
        }
    }

    def execute(self, data) -> Dict[str, Any]:
        """
        执行发送逻辑

        Inputs:
            - alert: 告警数据（已转换为钉钉格式）
            - webhook_url: Webhook URL
            - secret: 签名密钥
            - at_mobiles: @手机号列表
            - at_user_ids: @用户ID列表
            - is_at_all: @所有人
        """
        alert = data.get_one_of_inputs("alert", {})
        webhook_url = data.get_one_of_inputs("webhook_url")
        secret = data.get_one_of_inputs("secret", "")
        at_mobiles = data.get_one_of_inputs("at_mobiles", [])
        at_user_ids = data.get_one_of_inputs("at_user_ids", [])
        is_at_all = data.get_one_of_inputs("is_at_all", False)

        if not webhook_url:
            return self.Outputs(
                success=False,
                message="Dingtalk webhook URL not provided"
            )

        try:
            # 构建消息
            message = self.build_message(alert, at_mobiles, at_user_ids, is_at_all)

            # 发送消息
            url = webhook_url
            if secret:
                url = self.add_sign(webhook_url, secret)

            response = self.send_message(url, message)

            return self.Outputs(
                success=True,
                message="Alert sent to Dingtalk successfully",
                response=response
            )

        except Exception as e:
            return self.Outputs(
                success=False,
                message=f"Failed to send alert: {str(e)}",
                error=str(e)
            )

    def build_message(self, alert: Dict[str, Any], at_mobiles: list, at_user_ids: list, is_at_all: bool) -> Dict[str, Any]:
        """构建钉钉消息"""
        # 如果 alert 已经是钉钉格式，直接使用
        if "msgtype" in alert:
            message = alert
        else:
            # 否则转换
            message = {
                "msgtype": "text",
                "text": {
                    "content": str(alert)
                }
            }

        # 添加 @ 信息
        if at_mobiles or at_user_ids or is_at_all:
            message["at"] = {
                "atMobiles": at_mobiles,
                "atUserIds": at_user_ids,
                "isAtAll": is_at_all
            }

        return message

    def add_sign(self, webhook_url: str, secret: str) -> str:
        """添加签名"""
        timestamp = int(time.time() * 1000)
        secret_enc = secret.encode('utf-8')
        string_to_sign = f'{timestamp}\n{secret}'
        string_to_sign_enc = string_to_sign.encode('utf-8')

        hmac_code = hmac.new(secret_enc, string_to_sign_enc, digestmod=hashlib.sha256).digest()
        sign = base64.b64encode(hmac_code).decode('utf-8')

        return f"{webhook_url}&timestamp={timestamp}&sign={sign}"

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
