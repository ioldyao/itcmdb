"use client"

import React from "react"

import { Bell, ChevronDown, HelpCircle, Search, Settings, Share2 } from "lucide-react"
import { Input } from "@/components/ui/input"
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu"

export function Header() {
  return (
    <header className="h-14 bg-[#1e3a5f] text-white flex items-center px-4 justify-between">
      <div className="flex items-center gap-6">
        {/* Logo */}
        <div className="flex items-center gap-2">
          <div className="w-8 h-8 rounded-full bg-gradient-to-r from-cyan-400 to-blue-500 flex items-center justify-center">
            <span className="text-sm font-bold">M</span>
          </div>
          <span className="font-semibold">监控平台</span>
        </div>

        {/* Navigation */}
        <nav className="flex items-center gap-1">
          <DropdownMenu>
            <DropdownMenuTrigger className="flex items-center gap-1 px-3 py-2 hover:bg-white/10 rounded text-sm">
              常用 <ChevronDown className="w-4 h-4" />
            </DropdownMenuTrigger>
            <DropdownMenuContent>
              <DropdownMenuItem>告警事件</DropdownMenuItem>
              <DropdownMenuItem>仪表盘</DropdownMenuItem>
            </DropdownMenuContent>
          </DropdownMenu>
          <NavItem>首页</NavItem>
          <NavItem>仪表盘</NavItem>
          <NavItem>数据探索</NavItem>
          <NavItem active>告警事件</NavItem>
          <NavItem>观测场景</NavItem>
          <NavItem>配置</NavItem>
          <NavItem>...</NavItem>
        </nav>
      </div>

      <div className="flex items-center gap-4">
        {/* Search */}
        <div className="relative">
          <Input
            placeholder="全站搜索 Ctrl + k"
            className="w-48 h-8 bg-white/10 border-white/20 text-white placeholder:text-white/50 text-sm"
          />
          <Search className="absolute right-2 top-1/2 -translate-y-1/2 w-4 h-4 text-white/50" />
        </div>

        {/* Icons */}
        <div className="flex items-center gap-3">
          <button className="p-1 hover:bg-white/10 rounded">
            <Settings className="w-5 h-5" />
          </button>
          <button className="p-1 hover:bg-white/10 rounded">
            <Share2 className="w-5 h-5" />
          </button>
          <button className="p-1 hover:bg-white/10 rounded">
            <HelpCircle className="w-5 h-5" />
          </button>
        </div>

        {/* User */}
        <DropdownMenu>
          <DropdownMenuTrigger className="flex items-center gap-1 text-sm">
            123602315 <ChevronDown className="w-4 h-4" />
          </DropdownMenuTrigger>
          <DropdownMenuContent align="end">
            <DropdownMenuItem>个人设置</DropdownMenuItem>
            <DropdownMenuItem>退出登录</DropdownMenuItem>
          </DropdownMenuContent>
        </DropdownMenu>
      </div>
    </header>
  )
}

function NavItem({ children, active = false }: { children: React.ReactNode; active?: boolean }) {
  return (
    <button
      className={`px-3 py-2 text-sm relative ${
        active ? "text-white" : "text-white/80 hover:text-white"
      }`}
    >
      {children}
      {active && (
        <div className="absolute bottom-0 left-1/2 -translate-x-1/2 w-12 h-0.5 bg-white" />
      )}
    </button>
  )
}
