import { useState, useEffect } from 'react'
import { Card, Select, Button, Spin, Empty } from 'antd'
import { CameraOutlined, ReloadOutlined } from '@ant-design/icons'
import ReactECharts from 'echarts-for-react'
import { alertService } from '@/services/alertService'

const timeRangeMap: Record<string, { label: string; start: string; end: string }> = {
  '1h': { label: '1小时', start: 'now-1h', end: 'now' },
  '1d': { label: '1天', start: 'now-1d', end: 'now' },
  '1w': { label: '1周', start: 'now-1w', end: 'now' },
  '1M': { label: '1月', start: 'now-1M', end: 'now' },
}

export default function AlertAnalysis() {
  const [timeRange, setTimeRange] = useState('1w')
  const [loading, setLoading] = useState(false)
  const [chartData, setChartData] = useState<any>(null)

  const loadAnalytics = async () => {
    setLoading(true)
    try {
      const range = timeRangeMap[timeRange] || timeRangeMap['1w']
      const res = await alertService.getAlertAnalytics({
        start_time: range.start,
        end_time: range.end,
        group_by: ['status', 'severity'],
      })
      if (res.code === 0 && res.data) {
        setChartData(res.data)
      }
    } catch (error) {
      console.error('Failed to load analytics:', error)
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    loadAnalytics()
  }, [timeRange])

  const dates = chartData?.time_series?.dates || []
  const series = chartData?.time_series?.series || []

  const statusNameMap: Record<string, string> = {
    firing: '未恢复',
    acknowledged: '已确认',
    resolved: '已恢复',
    closed: '已关闭',
  }
  const statusColorMap: Record<string, string> = {
    firing: '#ff4d4f',
    acknowledged: '#fa8c16',
    resolved: '#52c41a',
    closed: '#d9d9d9',
  }

  const option = {
    tooltip: {
      trigger: 'axis',
      axisPointer: { type: 'shadow' },
    },
    legend: {
      data: series.map((s: any) => statusNameMap[s.name] || s.name),
      right: 0,
      top: 0,
      textStyle: { fontSize: 12 },
    },
    grid: {
      left: '3%',
      right: '4%',
      bottom: '15%',
      top: '15%',
      containLabel: true,
    },
    xAxis: {
      type: 'category',
      data: dates,
      axisLine: { show: false },
      axisTick: { show: false },
      axisLabel: { fontSize: 12 },
    },
    yAxis: {
      type: 'value',
      axisLine: { show: false },
      axisTick: { show: false },
      splitLine: { lineStyle: { type: 'dashed' } },
      axisLabel: { fontSize: 12 },
    },
    dataZoom: [
      {
        type: 'slider',
        show: true,
        xAxisIndex: [0],
        start: 0,
        end: 100,
        bottom: 5,
        height: 20,
        brushSelect: false,
      },
    ],
    series: series.map((s: any) => ({
      name: statusNameMap[s.name] || s.name,
      type: 'bar',
      stack: 'total',
      data: s.data,
      itemStyle: {
        color: statusColorMap[s.name] || '#999',
        borderRadius: [2, 2, 0, 0],
      },
      barWidth: '60%',
    })),
  }

  return (
    <Card bordered={false} className="dark:bg-bg-secondary dark:border-white/8">
      {/* 标题栏 */}
      <div className="flex items-center justify-between mb-4">
        <div className="flex items-center gap-4">
          <span className="text-base font-medium text-gray-900 dark:text-text-primary">告警趋势</span>
          <div className="flex items-center gap-2">
            <span className="text-sm text-gray-500 dark:text-text-secondary">汇聚周期</span>
            <Select
              value={timeRange}
              onChange={setTimeRange}
              className="w-24"
              size="small"
            >
              <Select.Option value="1h">1小时</Select.Option>
              <Select.Option value="1d">1天</Select.Option>
              <Select.Option value="1w">1周</Select.Option>
              <Select.Option value="1M">1月</Select.Option>
            </Select>
          </div>
        </div>
        <div className="flex gap-2">
          <Button
            icon={<ReloadOutlined />}
            size="small"
            type="text"
            onClick={loadAnalytics}
            loading={loading}
          />
          <Button
            icon={<CameraOutlined />}
            size="small"
            type="text"
          />
        </div>
      </div>

      {/* 图表 */}
      <div className="h-[400px]">
        {loading ? (
          <div className="flex justify-center items-center h-full">
            <Spin size="large" tip="加载中..." />
          </div>
        ) : dates.length > 0 ? (
          <ReactECharts option={option} style={{ height: '100%', width: '100%' }} />
        ) : (
          <Empty description="暂无数据" />
        )}
      </div>
    </Card>
  )
}
