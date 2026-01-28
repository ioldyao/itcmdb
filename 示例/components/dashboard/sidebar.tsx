"use client"

import React from "react"

import { useState } from "react"
import {
  AlertCircle,
  Bell,
  BellOff,
  CheckCircle,
  ChevronDown,
  ChevronRight,
  Eye,
  Star,
  User,
} from "lucide-react"
import { Checkbox } from "@/components/ui/checkbox"
import { Badge } from "@/components/ui/badge"
import { cn } from "@/lib/utils"

interface SidebarItemProps {
  icon: React.ReactNode
  label: string
  count: number
  active?: boolean
  iconColor?: string
}

function SidebarItem({ icon, label, count, active = false, iconColor }: SidebarItemProps) {
  return (
    <div
      className={cn(
        "flex items-center justify-between px-4 py-2 cursor-pointer hover:bg-muted/50",
        active && "bg-muted"
      )}
    >
      <div className="flex items-center gap-2">
        <span className={iconColor}>{icon}</span>
        <span className="text-sm">{label}</span>
      </div>
      <span className="text-sm text-muted-foreground">{count}</span>
    </div>
  )
}

interface FilterSectionProps {
  title: string
  children: React.ReactNode
  defaultOpen?: boolean
}

function FilterSection({ title, children, defaultOpen = true }: FilterSectionProps) {
  const [isOpen, setIsOpen] = useState(defaultOpen)

  return (
    <div className="border-b border-border">
      <button
        onClick={() => setIsOpen(!isOpen)}
        className="flex items-center gap-1 px-4 py-2 w-full text-left text-sm font-medium text-muted-foreground hover:bg-muted/50"
      >
        {isOpen ? <ChevronDown className="w-4 h-4" /> : <ChevronRight className="w-4 h-4" />}
        {title}
      </button>
      {isOpen && <div className="pb-2">{children}</div>}
    </div>
  )
}

interface FilterCheckboxProps {
  label: string
  count: number
  color?: string
}

function FilterCheckbox({ label, count, color }: FilterCheckboxProps) {
  return (
    <label className="flex items-center justify-between px-4 py-1.5 cursor-pointer hover:bg-muted/50">
      <div className="flex items-center gap-2">
        <Checkbox className="w-4 h-4" />
        {color && <div className={`w-1 h-4 rounded ${color}`} />}
        <span className={cn("text-sm", color?.includes("red") && "text-red-500", color?.includes("orange") && "text-orange-500", color?.includes("yellow") && "text-yellow-500")}>{label}</span>
      </div>
      <span className="text-sm text-muted-foreground">{count}</span>
    </label>
  )
}

export function Sidebar() {
  const [collapsed, setCollapsed] = useState(false)

  return (
    <aside className="w-60 border-r border-border bg-background flex flex-col h-full overflow-y-auto">
      {/* Alert Section */}
      <div className="border-b border-border">
        <div className="flex items-center justify-between px-4 py-3">
          <div className="flex items-center gap-2">
            <ChevronDown className="w-4 h-4 text-muted-foreground" />
            <span className="font-medium">告警</span>
          </div>
          <Badge className="bg-[#3b82f6] text-white hover:bg-[#3b82f6]">1284</Badge>
        </div>

        <SidebarItem
          icon={<User className="w-4 h-4" />}
          label="我负责的"
          count={0}
          iconColor="text-blue-500"
        />
        <SidebarItem
          icon={<Star className="w-4 h-4" />}
          label="我关注的"
          count={0}
          iconColor="text-green-500"
        />
        <SidebarItem
          icon={<Bell className="w-4 h-4" />}
          label="我收到的"
          count={0}
        />
        <SidebarItem
          icon={<AlertCircle className="w-4 h-4" />}
          label="未恢复"
          count={230}
          iconColor="text-red-500"
        />
        <SidebarItem
          icon={<BellOff className="w-4 h-4" />}
          label="未恢复(已屏蔽)"
          count={6}
        />
        <SidebarItem
          icon={<CheckCircle className="w-4 h-4" />}
          label="已恢复"
          count={1045}
          iconColor="text-green-500"
        />
      </div>

      {/* Processing Records */}
      <div className="border-b border-border">
        <div className="flex items-center gap-2 px-4 py-3">
          <ChevronRight className="w-4 h-4 text-muted-foreground" />
          <span className="font-medium">处理记录</span>
          <span className="text-sm text-muted-foreground ml-auto">5509</span>
        </div>
      </div>

      {/* Advanced Filters */}
      <div className="px-4 py-3 border-b border-border">
        <span className="text-sm font-medium text-muted-foreground">高级筛选</span>
      </div>

      <FilterSection title="级别">
        <FilterCheckbox label="致命" count={15} color="bg-red-500" />
        <FilterCheckbox label="预警" count={1261} color="bg-orange-500" />
        <FilterCheckbox label="提醒" count={8} color="bg-yellow-500" />
      </FilterSection>

      <FilterSection title="处理阶段">
        <FilterCheckbox label="已通知" count={224} />
        <FilterCheckbox label="已确认" count={3} />
        <FilterCheckbox label="已屏蔽" count={369} />
        <FilterCheckbox label="已流控" count={0} />
      </FilterSection>

      <FilterSection title="数据类型">
        <FilterCheckbox label="时序数据" count={1283} />
        <FilterCheckbox label="事件" count={1} />
        <FilterCheckbox label="日志" count={0} />
      </FilterSection>

      <FilterSection title="分类" defaultOpen={false}>
        <div className="px-4 py-2 text-sm text-muted-foreground">加载中...</div>
      </FilterSection>
    </aside>
  )
}
