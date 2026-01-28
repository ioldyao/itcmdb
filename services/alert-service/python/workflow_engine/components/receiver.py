"""
告警接收组件
接收来自外部系统的告警数据
"""

from typing import Dict, Any
from pipeline.component.framework.component import Component


class AlertReceiverComponent(Component):
    """告警接收组件"""

    name = "告警接收组件"
    code = "alert_receiver"
    form_schema = {
        "type": "object",
        "properties": {
            "source_type": {
                "type": "string",
                "enum": ["alertmanager", "prometheus", "victoriametrics", "custom"],
                "title": "来源类型"
            }
        }
    }

    def execute(self, data) -> Dict[str, Any]:
        """
        执行接收逻辑

        Inputs:
            - alert: 告警数据
            - source_type: 来源类型
        """
        alert = data.get_one_of_inputs("alert")
        source_type = data.get_one_of_inputs("source_type", "custom")

        # 标准化告警格式
        normalized_alert = self.normalize_alert(alert, source_type)

        return self.Outputs(
            success=True,
            alert=normalized_alert,
            original_alert=alert
        )

    def normalize_alert(self, alert: Dict[str, Any], source_type: str) -> Dict[str, Any]:
        """标准化告警格式"""
        if source_type == "alertmanager":
            return self.normalize_alertmanager(alert)
        elif source_type == "prometheus":
            return self.normalize_prometheus(alert)
        elif source_type == "victoriametrics":
            return self.normalize_victoriametrics(alert)
        else:
            return alert

    def normalize_alertmanager(self, alert: Dict[str, Any]) -> Dict[str, Any]:
        """转换 Alertmanager 格式"""
        return {
            "status": alert.get("status", "firing"),
            "labels": alert.get("labels", {}),
            "annotations": alert.get("annotations", {}),
            "startsAt": alert.get("startsAt"),
            "endsAt": alert.get("endsAt"),
            "generatorURL": alert.get("generatorURL"),
            "fingerprint": alert.get("fingerprint"),
        }

    def normalize_prometheus(self, alert: Dict[str, Any]) -> Dict[str, Any]:
        """转换 Prometheus 格式"""
        return self.normalize_alertmanager(alert)

    def normalize_victoriametrics(self, alert: Dict[str, Any]) -> Dict[str, Any]:
        """转换 VictoriaMetrics 格式"""
        return self.normalize_alertmanager(alert)

    def outputs(self) -> list:
        return [
            {
                "name": "success",
                "type": "boolean",
                "description": "是否成功"
            },
            {
                "name": "alert",
                "type": "object",
                "description": "标准化后的告警"
            },
            {
                "name": "original_alert",
                "type": "object",
                "description": "原始告警数据"
            }
        ]
