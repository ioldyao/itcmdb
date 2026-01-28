"""
告警过滤组件
根据条件过滤告警
"""

from typing import Dict, Any, List
from pipeline.component.framework.component import Component


class AlertFilterComponent(Component):
    """告警过滤组件"""

    name = "告警过滤组件"
    code = "alert_filter"
    form_schema = {
        "type": "object",
        "properties": {
            "conditions": {
                "type": "array",
                "title": "过滤条件",
                "items": {
                    "type": "object",
                    "properties": {
                        "field": {"type": "string", "title": "字段"},
                        "operator": {
                            "type": "string",
                            "enum": ["==", "!=", ">", "<", ">=", "<=", "in", "contains", "regex"],
                            "title": "操作符"
                        },
                        "value": {"title": "值"}
                    }
                }
            },
            "match_mode": {
                "type": "string",
                "enum": ["all", "any"],
                "default": "all",
                "title": "匹配模式"
            }
        }
    }

    def execute(self, data) -> Dict[str, Any]:
        """
        执行过滤逻辑

        Inputs:
            - alert: 告警数据
            - conditions: 过滤条件
            - match_mode: 匹配模式（all/any）
        """
        alert = data.get_one_of_inputs("alert", {})
        conditions = data.get_one_of_inputs("conditions", [])
        match_mode = data.get_one_of_inputs("match_mode", "all")

        if not conditions:
            # 没有条件，通过所有告警
            return self.Outputs(
                is_filtered=False,
                alert=alert,
                message="No filter conditions"
            )

        # 执行过滤
        matched = self.evaluate_conditions(alert, conditions, match_mode)

        if matched:
            return self.Outputs(
                is_filtered=False,
                alert=alert,
                message="Alert passed filter"
            )
        else:
            return self.Outputs(
                is_filtered=True,
                alert=None,
                message="Alert was filtered out"
            )

    def evaluate_conditions(self, alert: Dict[str, Any], conditions: List[Dict], match_mode: str) -> bool:
        """评估过滤条件"""
        results = []

        for condition in conditions:
            field = condition["field"]
            operator = condition["operator"]
            expected_value = condition["value"]

            # 获取字段值
            actual_value = self.get_field_value(alert, field)

            # 比较值
            result = self.compare_values(actual_value, operator, expected_value)
            results.append(result)

        # 根据匹配模式返回结果
        if match_mode == "all":
            return all(results)
        else:  # any
            return any(results)

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

    def compare_values(self, actual: Any, operator: str, expected: Any) -> bool:
        """比较值"""
        try:
            if operator == "==":
                return actual == expected
            elif operator == "!=":
                return actual != expected
            elif operator == ">":
                return float(actual) > float(expected)
            elif operator == "<":
                return float(actual) < float(expected)
            elif operator == ">=":
                return float(actual) >= float(expected)
            elif operator == "<=":
                return float(actual) <= float(expected)
            elif operator == "in":
                return actual in expected if isinstance(expected, list) else actual in str(expected).split(",")
            elif operator == "contains":
                return str(expected) in str(actual)
            elif operator == "regex":
                import re
                return bool(re.match(str(expected), str(actual)))
            else:
                return False
        except (TypeError, ValueError):
            return False

    def outputs(self) -> list:
        return [
            {
                "name": "is_filtered",
                "type": "boolean",
                "description": "是否被过滤"
            },
            {
                "name": "alert",
                "type": "object",
                "description": "过滤后的告警"
            },
            {
                "name": "message",
                "type": "string",
                "description": "过滤结果消息"
            }
        ]
