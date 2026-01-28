"use client"

import { useState } from "react"
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table"
import { Checkbox } from "@/components/ui/checkbox"
import { Badge } from "@/components/ui/badge"
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs"
import { Input } from "@/components/ui/input"
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select"
import { AlertCircle, ChevronDown, Download, Filter, Settings, Star } from "lucide-react"
import { cn } from "@/lib/utils"

interface AlertData {
  id: string
  spaceName: string
  alertName: string
  category: string
  metric: string
  metricValue: number
  stage: string
  status: "未恢复" | "已恢复"
}

const alertsData: AlertData[] = [
  {
    id: "1769562032614...",
    spaceName: "--",
    alertName: "在线人数告警",
    category: "用户体验-业务应用",
    metric: "在线",
    metricValue: 4,
    stage: "已屏蔽",
    status: "未恢复",
  },
  {
    id: "1769529059614...",
    spaceName: "--",
    alertName: "<h1>cgg</h1>*}}{{2*2-1}}}${22*2-1}",
    category: "主机&云平台-操作系统",
    metric: "Agent心跳丢失",
    metricValue: 54,
    stage: "已通知",
    status: "未恢复",
  },
  {
    id: "1768871964606...",
    spaceName: "--",
    alertName: "测试2",
    category: "主机&云平台-操作系统",
    metric: "CPU使用率",
    metricValue: 11,
    stage: "已屏蔽",
    status: "未恢复",
  },
  {
    id: "1768871964606...",
    spaceName: "--",
    alertName: "测试2",
    category: "主机&云平台-操作系统",
    metric: "CPU使用率",
    metricValue: 11,
    stage: "已通知",
    status: "未恢复",
  },
  {
    id: "1765935553569...",
    spaceName: "--",
    alertName: "121212",
    category: "用户体验-服务拨测",
    metric: "HTTP 期望响应码",
    metricValue: 59,
    stage: "已通知",
    status: "未恢复",
  },
]

export function AlertTable() {
  const [selectedAlerts, setSelectedAlerts] = useState<string[]>([])

  return (
    <div className="flex flex-col gap-4">
      {/* Filter Bar */}
      <div className="flex items-center gap-4">
        <div className="flex items-center gap-2">
          <span className="text-sm text-muted-foreground">空间筛选</span>
          <span className="text-sm text-[#3b82f6]">空间筛选：</span>
          <Select defaultValue="all">
            <SelectTrigger className="w-48 h-8">
              <SelectValue placeholder="-我有权限的空间-" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="all">-我有权限的空间-</SelectItem>
              <SelectItem value="space1">空间1</SelectItem>
              <SelectItem value="space2">空间2</SelectItem>
            </SelectContent>
          </Select>
        </div>

        <div className="flex-1 flex items-center gap-2">
          <div className="relative flex-1 max-w-md">
            <Filter className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-muted-foreground" />
            <Input
              placeholder="输入搜索条件"
              className="pl-9 h-8"
            />
          </div>
          <button className="flex items-center gap-1 px-3 py-1.5 text-sm text-muted-foreground hover:bg-muted rounded border border-border">
            <Star className="w-4 h-4" />
            收藏
          </button>
          <button className="p-1.5 hover:bg-muted rounded border border-border">
            <Download className="w-4 h-4 text-muted-foreground" />
          </button>
        </div>
      </div>

      {/* Alert Notice */}
      <div className="flex items-center gap-2 px-4 py-2 bg-blue-50 text-blue-700 rounded border border-blue-200">
        <AlertCircle className="w-4 h-4" />
        <span className="text-sm">当前有 3 个未恢复告警的通知人是空的，</span>
        <button className="text-sm text-blue-600 hover:underline">查看</button>
      </div>

      {/* Tabs */}
      <Tabs defaultValue="list">
        <TabsList className="bg-transparent border-b border-border rounded-none w-full justify-start gap-6 h-auto p-0">
          <TabsTrigger
            value="list"
            className="rounded-none border-b-2 border-transparent data-[state=active]:border-[#3b82f6] data-[state=active]:bg-transparent data-[state=active]:shadow-none px-0 pb-2"
          >
            告警列表
          </TabsTrigger>
          <TabsTrigger
            value="analysis"
            className="rounded-none border-b-2 border-transparent data-[state=active]:border-[#3b82f6] data-[state=active]:bg-transparent data-[state=active]:shadow-none px-0 pb-2"
          >
            告警分析
          </TabsTrigger>
        </TabsList>

        <TabsContent value="list" className="mt-4">
          <div className="border border-border rounded-lg overflow-hidden">
            <Table>
              <TableHeader>
                <TableRow className="bg-muted/30">
                  <TableHead className="w-10">
                    <Checkbox />
                  </TableHead>
                  <TableHead>告警ID</TableHead>
                  <TableHead>空间名</TableHead>
                  <TableHead>告警名称</TableHead>
                  <TableHead>分类</TableHead>
                  <TableHead>
                    <div className="flex items-center gap-1">
                      告警指标
                      <ChevronDown className="w-3 h-3" />
                    </div>
                  </TableHead>
                  <TableHead>关</TableHead>
                  <TableHead>处理阶段</TableHead>
                  <TableHead>
                    <div className="flex items-center gap-1">
                      状态
                      <ChevronDown className="w-3 h-3" />
                    </div>
                  </TableHead>
                  <TableHead className="w-10">
                    <Settings className="w-4 h-4 text-muted-foreground" />
                  </TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {alertsData.map((alert, index) => (
                  <TableRow key={index}>
                    <TableCell>
                      <Checkbox />
                    </TableCell>
                    <TableCell>
                      <div className="flex items-center gap-1">
                        <div className="w-1 h-4 bg-red-500 rounded" />
                        <span className="text-[#3b82f6] hover:underline cursor-pointer text-sm">
                          {alert.id}
                        </span>
                      </div>
                    </TableCell>
                    <TableCell className="text-sm">{alert.spaceName}</TableCell>
                    <TableCell className="text-sm max-w-[200px] truncate">
                      {alert.alertName}
                    </TableCell>
                    <TableCell className="text-sm">{alert.category}</TableCell>
                    <TableCell>
                      <Badge variant="secondary" className="text-xs font-normal bg-muted">
                        {alert.metric}
                      </Badge>
                    </TableCell>
                    <TableCell>
                      <span className="text-sm text-[#3b82f6]">{alert.metricValue}</span>
                    </TableCell>
                    <TableCell className="text-sm">{alert.stage}</TableCell>
                    <TableCell>
                      <span
                        className={cn(
                          "text-sm",
                          alert.status === "未恢复" ? "text-red-500" : "text-green-500"
                        )}
                      >
                        {alert.status}
                      </span>
                    </TableCell>
                    <TableCell></TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </div>
        </TabsContent>

        <TabsContent value="analysis" className="mt-4">
          <div className="text-center py-12 text-muted-foreground">
            告警分析功能开发中...
          </div>
        </TabsContent>
      </Tabs>
    </div>
  )
}
