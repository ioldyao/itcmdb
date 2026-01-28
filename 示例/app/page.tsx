"use client"

import { useState } from "react"
import { Header } from "@/components/dashboard/header"
import { Sidebar } from "@/components/dashboard/sidebar"
import { AlertChart } from "@/components/dashboard/alert-chart"
import { AlertTable } from "@/components/dashboard/alert-table"
import {
  ChevronLeft,
  ChevronRight,
  Clock,
  ExternalLink,
  Maximize2,
  RefreshCw,
} from "lucide-react"
import { Button } from "@/components/ui/button"

export default function MonitoringDashboard() {
  const [sidebarCollapsed, setSidebarCollapsed] = useState(false)

  return (
    <div className="min-h-screen flex flex-col bg-muted/30">
      <Header />

      <div className="flex flex-1 overflow-hidden">
        {/* Sidebar */}
        <div className={`relative ${sidebarCollapsed ? "w-0" : "w-60"} transition-all duration-300`}>
          {!sidebarCollapsed && <Sidebar />}
          <button
            onClick={() => setSidebarCollapsed(!sidebarCollapsed)}
            className="absolute -right-3 top-1/2 -translate-y-1/2 z-10 w-6 h-12 bg-background border border-border rounded-r flex items-center justify-center hover:bg-muted"
          >
            {sidebarCollapsed ? (
              <ChevronRight className="w-4 h-4" />
            ) : (
              <ChevronLeft className="w-4 h-4" />
            )}
          </button>
        </div>

        {/* Main Content */}
        <main className="flex-1 overflow-y-auto p-6">
          {/* Page Header */}
          <div className="flex items-center justify-between mb-6">
            <h1 className="text-xl font-semibold">告警</h1>
            <div className="flex items-center gap-3">
              <Button variant="ghost" size="icon" className="w-8 h-8">
                <ChevronLeft className="w-4 h-4" />
              </Button>
              <div className="flex items-center gap-1 px-3 py-1.5 bg-background border border-border rounded text-sm">
                <Clock className="w-4 h-4 text-muted-foreground" />
                <span>近 7 天</span>
              </div>
              <Button variant="ghost" size="icon" className="w-8 h-8">
                <ChevronRight className="w-4 h-4" />
              </Button>
              <div className="flex items-center gap-1 text-sm text-muted-foreground">
                <Clock className="w-4 h-4" />
                <span>5m</span>
              </div>
              <Button variant="ghost" size="icon" className="w-8 h-8">
                <RefreshCw className="w-4 h-4" />
              </Button>
              <Button variant="ghost" size="icon" className="w-8 h-8">
                <Maximize2 className="w-4 h-4" />
              </Button>
            </div>
          </div>

          {/* Alert Trend Chart */}
          <div className="mb-6">
            <AlertChart />
          </div>

          {/* Alert Table */}
          <AlertTable />
        </main>
      </div>
    </div>
  )
}
