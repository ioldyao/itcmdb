"""
FastAPI 应用入口
工作流引擎 REST API
"""

from fastapi import FastAPI, HTTPException, status
from fastapi.middleware.cors import CORSMiddleware
from pydantic import BaseModel, Field
from typing import Dict, Any, Optional, List
import uuid
from datetime import datetime

from bamboo_engine import api as engine_api
from bamboo_engine.builder import (
    build_tree,
    EmptyStartEvent,
    EmptyEndEvent,
    ServiceActivity,
    ParallelGateway,
    ConvergeGateway,
    ExclusiveGateway,
    Data,
)

# 创建 FastAPI 应用
app = FastAPI(
    title="ITCMDB Workflow Engine",
    description="基于 Bamboo-Engine 的告警集成工作流引擎",
    version="1.0.0",
)

# 配置 CORS
app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

# 内存存储（生产环境应使用数据库）
workflows_store: Dict[str, Dict] = {}
webhooks_store: Dict[str, Dict] = {}


# ==================== 请求/响应模型 ====================

class WorkflowExecuteRequest(BaseModel):
    """工作流执行请求"""
    pipeline: Dict[str, Any]
    data: Optional[Dict[str, Any]] = {}


class WorkflowExecuteResponse(BaseModel):
    """工作流执行响应"""
    success: bool
    pipeline_id: Optional[str] = None
    message: Optional[str] = None


class WebhookReceiveRequest(BaseModel):
    """Webhook 接收请求"""
    alert_data: Dict[str, Any]


class WorkflowStatusResponse(BaseModel):
    """工作流状态响应"""
    success: bool
    states: Optional[Dict[str, Any]] = None
    message: Optional[str] = None


class WorkflowCreateRequest(BaseModel):
    """创建工作流请求"""
    name: str
    description: Optional[str] = None
    direction: str = Field(..., pattern="^(inbound|outbound)$")
    workflow_type: str = Field(..., pattern="^(alertmanager|prometheus|victoriametrics|workflow)$")
    pipeline: Dict[str, Any]
    enabled: bool = True


class WebhookCreateResponse(BaseModel):
    """Webhook 创建响应"""
    success: bool
    webhook_id: Optional[str] = None
    webhook_token: Optional[str] = None
    webhook_url: Optional[str] = None
    message: Optional[str] = None


# ==================== 辅助函数 ====================

def get_workflow_by_token(token: str) -> Optional[Dict]:
    """根据 token 获取工作流配置"""
    for webhook_id, webhook_config in webhooks_store.items():
        if webhook_config.get("webhook_token") == token:
            workflow_id = webhook_config.get("workflow_id")
            return workflows_store.get(workflow_id)
    return None


def generate_webhook_token() -> str:
    """生成 Webhook Token"""
    return f"wh_{uuid.uuid4().hex[:16]}"


# ==================== API 端点 ====================

@app.get("/")
async def root():
    """健康检查"""
    return {
        "service": "ITCMDB Workflow Engine",
        "version": "1.0.0",
        "status": "running",
        "timestamp": datetime.now().isoformat()
    }


@app.post("/api/v1/workflow/execute", response_model=WorkflowExecuteResponse)
async def execute_workflow(req: WorkflowExecuteRequest):
    """
    执行工作流

    - **pipeline**: Bamboo-Engine Pipeline 定义
    - **data**: Pipeline 输入数据
    """
    try:
        # 模拟执行（实际需要 BambooDjangoRuntime）
        pipeline_id = req.pipeline.get("id", f"pipeline_{uuid.uuid4().hex}")

        # TODO: 实际执行
        # from pipeline.eri.runtime import BambooDjangoRuntime
        # runtime = BambooDjangoRuntime()
        # result = engine_api.run_pipeline(
        #     runtime=runtime,
        #     pipeline=req.pipeline,
        #     root_pipeline_data=req.data
        # )

        # 保存执行记录
        workflows_store[pipeline_id] = {
            "id": pipeline_id,
            "pipeline": req.pipeline,
            "data": req.data,
            "status": "running",
            "created_at": datetime.now().isoformat()
        }

        return WorkflowExecuteResponse(
            success=True,
            pipeline_id=pipeline_id,
            message="Workflow started successfully"
        )

    except Exception as e:
        raise HTTPException(
            status_code=status.HTTP_400_BAD_REQUEST,
            detail=f"Failed to execute workflow: {str(e)}"
        )


@app.post("/api/v1/webhooks/{token}", response_model=WorkflowExecuteResponse)
async def receive_webhook_alert(token: str, req: WebhookReceiveRequest):
    """
    接收 Webhook 告警并触发工作流

    - **token**: Webhook Token
    - **alert_data**: 告警数据
    """
    try:
        # 1. 根据 token 查找对应的工作流
        workflow = get_workflow_by_token(token)
        if not workflow:
            raise HTTPException(
                status_code=status.HTTP_404_NOT_FOUND,
                detail=f"Webhook token not found: {token}"
            )

        # 2. 获取 Pipeline 配置
        pipeline = workflow.get("pipeline")
        if not pipeline:
            raise HTTPException(
                status_code=status.HTTP_400_BAD_REQUEST,
                detail="Workflow pipeline not configured"
            )

        # 3. 执行工作流
        pipeline_id = f"exec_{uuid.uuid4().hex}"

        # TODO: 实际执行
        # runtime = BambooDjangoRuntime()
        # result = engine_api.run_pipeline(
        #     runtime=runtime,
        #     pipeline=pipeline,
        #     root_pipeline_data={"alert": req.alert_data}
        # )

        # 保存执行记录
        workflows_store[pipeline_id] = {
            "id": pipeline_id,
            "pipeline": pipeline,
            "alert_data": req.alert_data,
            "status": "running",
            "created_at": datetime.now().isoformat()
        }

        return WorkflowExecuteResponse(
            success=True,
            pipeline_id=pipeline_id,
            message="Alert received and workflow started"
        )

    except HTTPException:
        raise
    except Exception as e:
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail=f"Failed to process webhook: {str(e)}"
        )


@app.get("/api/v1/workflow/{pipeline_id}/status", response_model=WorkflowStatusResponse)
async def get_workflow_status(pipeline_id: str):
    """
    获取工作流执行状态

    - **pipeline_id**: Pipeline ID
    """
    try:
        workflow = workflows_store.get(pipeline_id)
        if not workflow:
            return WorkflowStatusResponse(
                success=False,
                message=f"Workflow not found: {pipeline_id}"
            )

        # TODO: 实际状态查询
        # runtime = BambooDjangoRuntime()
        # result = engine_api.get_pipeline_states(
        #     runtime=runtime,
        #     root_id=pipeline_id,
        #     flat_children=True
        # )

        return WorkflowStatusResponse(
            success=True,
            states={
                "id": workflow["id"],
                "status": workflow.get("status", "unknown"),
                "created_at": workflow.get("created_at")
            }
        )

    except Exception as e:
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail=f"Failed to get workflow status: {str(e)}"
        )


@app.post("/api/v1/workflow/{pipeline_id}/pause")
async def pause_workflow(pipeline_id: str):
    """
    暂停工作流

    - **pipeline_id**: Pipeline ID
    """
    try:
        workflow = workflows_store.get(pipeline_id)
        if not workflow:
            raise HTTPException(
                status_code=status.HTTP_404_NOT_FOUND,
                detail=f"Workflow not found: {pipeline_id}"
            )

        # TODO: 实际暂停
        # runtime = BambooDjangoRuntime()
        # result = engine_api.pause_pipeline(runtime=runtime, pipeline_id=pipeline_id)

        workflow["status"] = "paused"

        return {"success": True, "message": "Workflow paused"}

    except HTTPException:
        raise
    except Exception as e:
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail=f"Failed to pause workflow: {str(e)}"
        )


@app.post("/api/v1/workflow/{pipeline_id}/resume")
async def resume_workflow(pipeline_id: str):
    """
    恢复工作流

    - **pipeline_id**: Pipeline ID
    """
    try:
        workflow = workflows_store.get(pipeline_id)
        if not workflow:
            raise HTTPException(
                status_code=status.HTTP_404_NOT_FOUND,
                detail=f"Workflow not found: {pipeline_id}"
            )

        # TODO: 实际恢复
        # runtime = BambooDjangoRuntime()
        # result = engine_api.resume_pipeline(runtime=runtime, pipeline_id=pipeline_id)

        workflow["status"] = "running"

        return {"success": True, "message": "Workflow resumed"}

    except HTTPException:
        raise
    except Exception as e:
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail=f"Failed to resume workflow: {str(e)}"
        )


@app.post("/api/v1/workflow", response_model=WebhookCreateResponse)
async def create_workflow(req: WorkflowCreateRequest):
    """
    创建工作流

    - **name**: 工作流名称
    - **description**: 工作流描述
    - **direction**: 方向（inbound/outbound）
    - **workflow_type**: 类型
    - **pipeline**: Pipeline 定义
    - **enabled**: 是否启用
    """
    try:
        workflow_id = f"wf_{uuid.uuid4().hex}"

        # 保存工作流
        workflows_store[workflow_id] = {
            "id": workflow_id,
            "name": req.name,
            "description": req.description,
            "direction": req.direction,
            "type": req.workflow_type,
            "pipeline": req.pipeline,
            "enabled": req.enabled,
            "created_at": datetime.now().isoformat()
        }

        # 如果是 inbound 类型，生成 webhook token
        if req.direction == "inbound":
            webhook_token = generate_webhook_token()
            webhook_id = f"wh_{uuid.uuid4().hex}"

            webhooks_store[webhook_id] = {
                "id": webhook_id,
                "workflow_id": workflow_id,
                "webhook_token": webhook_token,
                "enabled": req.enabled,
                "created_at": datetime.now().isoformat()
            }

            # 生成 Webhook URL
            base_url = "http://localhost:8000"  # TODO: 从配置读取
            webhook_url = f"{base_url}/api/v1/webhooks/{webhook_token}"

            return WebhookCreateResponse(
                success=True,
                webhook_id=webhook_id,
                webhook_token=webhook_token,
                webhook_url=webhook_url,
                message="Workflow created with webhook"
            )
        else:
            return WebhookCreateResponse(
                success=True,
                webhook_id=workflow_id,
                message="Workflow created successfully"
            )

    except Exception as e:
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail=f"Failed to create workflow: {str(e)}"
        )


@app.get("/api/v1/workflows")
async def list_workflows():
    """获取工作流列表"""
    return {
        "success": True,
        "workflows": list(workflows_store.values())
    }


@app.get("/api/v1/webhooks")
async def list_webhooks():
    """获取 Webhook 列表"""
    return {
        "success": True,
        "webhooks": list(webhooks_store.values())
    }


@app.get("/api/v1/webhooks/{webhook_id}")
async def get_webhook(webhook_id: str):
    """获取 Webhook 详情"""
    webhook = webhooks_store.get(webhook_id)
    if not webhook:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail=f"Webhook not found: {webhook_id}"
        )

    # 获取关联的工作流
    workflow_id = webhook.get("workflow_id")
    workflow = workflows_store.get(workflow_id)

    return {
        "success": True,
        "webhook": webhook,
        "workflow": workflow
    }


if __name__ == "__main__":
    import uvicorn
    uvicorn.run(app, host="0.0.0.0", port=8000, reload=True)
