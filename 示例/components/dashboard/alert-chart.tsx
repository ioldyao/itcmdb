"use client"

import {
  Bar,
  BarChart,
  ResponsiveContainer,
  XAxis,
  YAxis,
  Tooltip,
  Legend,
} from "recharts"
import { ChevronDown, Camera } from "lucide-react"
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select"

const chartData = [
  { date: "01-20", unrecovered: 320, recovered: 280, closed: 150 },
  { date: "01-21", unrecovered: 350, recovered: 300, closed: 180 },
  { date: "01-22", unrecovered: 280, recovered: 260, closed: 140 },
  { date: "01-23", unrecovered: 320, recovered: 290, closed: 160 },
  { date: "01-24", unrecovered: 290, recovered: 270, closed: 150 },
  { date: "01-25", unrecovered: 310, recovered: 280, closed: 170 },
  { date: "01-26", unrecovered: 340, recovered: 300, closed: 180 },
  { date: "01-27", unrecovered: 300, recovered: 260, closed: 140 },
]

export function AlertChart() {
  return (
    <div className="bg-background border border-border rounded-lg p-4">
      <div className="flex items-center justify-between mb-4">
        <div className="flex items-center gap-4">
          <div className="flex items-center gap-2">
            <ChevronDown className="w-4 h-4 text-muted-foreground" />
            <span className="font-medium">告警趋势</span>
          </div>
          <div className="flex items-center gap-2 text-sm">
            <span className="text-muted-foreground">汇聚周期</span>
            <Select defaultValue="auto">
              <SelectTrigger className="h-7 w-20">
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="auto">Auto</SelectItem>
                <SelectItem value="1h">1小时</SelectItem>
                <SelectItem value="1d">1天</SelectItem>
              </SelectContent>
            </Select>
          </div>
        </div>
        <button className="p-1 hover:bg-muted rounded">
          <Camera className="w-4 h-4 text-muted-foreground" />
        </button>
      </div>

      <div className="h-64">
        <ResponsiveContainer width="100%" height="100%">
          <BarChart data={chartData} barGap={0}>
            <XAxis
              dataKey="date"
              tick={{ fontSize: 12 }}
              tickLine={false}
              axisLine={false}
            />
            <YAxis
              tick={{ fontSize: 12 }}
              tickLine={false}
              axisLine={false}
              domain={[0, 400]}
              ticks={[0, 100, 200, 300, 400]}
            />
            <Tooltip
              contentStyle={{
                backgroundColor: "hsl(var(--background))",
                border: "1px solid hsl(var(--border))",
                borderRadius: "6px",
                fontSize: "12px",
              }}
            />
            <Legend
              align="right"
              verticalAlign="bottom"
              iconType="square"
              wrapperStyle={{ fontSize: "12px", paddingTop: "8px" }}
            />
            <Bar
              dataKey="unrecovered"
              name="未恢复"
              fill="#f87171"
              radius={[2, 2, 0, 0]}
            />
            <Bar
              dataKey="recovered"
              name="已恢复"
              fill="#86efac"
              radius={[2, 2, 0, 0]}
            />
            <Bar
              dataKey="closed"
              name="已关闭"
              fill="#d1d5db"
              radius={[2, 2, 0, 0]}
            />
          </BarChart>
        </ResponsiveContainer>
      </div>
    </div>
  )
}
