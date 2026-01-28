"""
Alertmanager 发送组件
推送告警到 Alertmanager
"""

import httpx
from typing import Dict, Any
from pipeline.component.framework.component import Component


class AlertmanagerSenderComponent(Component):
    """Alertmanager 发送组件"""

    name = "Alertmanager 发送组件"
    code = "sender_alertmanager"
    form_schema = {
        "type": "object",
        "properties": {
            "url": {
                "type": "string",
                "title": "Alertmanager URL",
                "format": "uri"
            },
            "timeout": {
                "type": "integer",
                "title": "超时时间（秒）",
                "default": 30
            }
        }
    }

    def execute(self, data) -> Dict[str, Any]:
        """
        执行发送逻辑

        Inputs:
            - alert: 告警数据
            - url: Alertmanager URL
            - timeout: 超时时间
        """
        alert = data.get_one_of_inputs("alert", {})
        url = data.get_one_of_inputs("url")
        timeout = data.get_one_of_inputs("timeout", 30)

        if not url:
            return self.Outputs(
                success=False,
                message="Alertmanager URL not provided"
            )

        try:
            # 发送到 Alertmanager
            response = self.send_to_alertmanager(url, alert, timeout)

            return self.Outputs(
                success=True,
                message="Alert sent to Alertmanager successfully",
                response=response
            )

        except Exception as e:
            return self.Outputs(
                success=False,
                message=f"Failed to send alert: {str(e)}",
                error=str(e)
            )

    def send_to_alertmanager(self, url: str, alert: Dict[str, Any], timeout: int) -> Dict[str, Any]:
        """发送告警到 Alertmanager"""
        # Alertmanager API 格式
        payload = [alert]

        with httpx.Client(timeout=timeout) as client:
            response = client.post(url, json=payload)
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
