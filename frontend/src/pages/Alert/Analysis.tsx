import { useState } from 'react'
import { Card, Select, Button } from 'antd'
import { CameraOutlined } from '@ant-design/icons'
import ReactECharts from 'echarts-for-react'
import dayjs from 'dayjs'

interface AlertAnalysisProps {
  statistics?: any
}

export default function AlertAnalysis({ }: AlertAnalysisProps) {
  const [timeRange, setTimeRange] = useState('1d')

  // 模拟数据 - 实际应从API获取
  const generateChartData = () => {
    const data = []
    const days = 7
    for (let i = days - 1; i >= 0; i--) {
      const date = dayjs().subtract(i, 'day').format('MM-DD')
      data.push({
        date,
        unrecovered: Math.floor(Math.random() * 100) + 200,
        recovered: Math.floor(Math.random() * 100) + 150,
        closed: Math.floor(Math.random() * 50) + 50,
      })
    }
    return data
  }

  const chartData = generateChartData()

  const option = {
    tooltip: {
      trigger: 'axis',
      axisPointer: {
        type: 'shadow'
      }
    },
    legend: {
      data: ['未恢复', '已恢复', '已关闭'],
      right: 0,
      top: 0,
      textStyle: {
        fontSize: 12
      }
    },
    grid: {
      left: '3%',
      right: '4%',
      bottom: '15%',
      top: '15%',
      containLabel: true
    },
    xAxis: {
      type: 'category',
      data: chartData.map(d => d.date),
      axisLine: {
        show: false
      },
      axisTick: {
        show: false
      },
      axisLabel: {
        fontSize: 12
      }
    },
    yAxis: {
      type: 'value',
      axisLine: {
        show: false
      },
      axisTick: {
        show: false
      },
      splitLine: {
        lineStyle: {
          type: 'dashed'
        }
      },
      axisLabel: {
        fontSize: 12
      }
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
        brushSelect: false
      }
    ],
    series: [
      {
        name: '未恢复',
        type: 'bar',
        stack: 'total',
        data: chartData.map(d => d.unrecovered),
        itemStyle: {
          color: '#ff4d4f',
          borderRadius: [2, 2, 0, 0]
        },
        barWidth: '60%'
      },
      {
        name: '已恢复',
        type: 'bar',
        stack: 'total',
        data: chartData.map(d => d.recovered),
        itemStyle: {
          color: '#52c41a',
          borderRadius: [2, 2, 0, 0]
        },
        barWidth: '60%'
      },
      {
        name: '已关闭',
        type: 'bar',
        stack: 'total',
        data: chartData.map(d => d.closed),
        itemStyle: {
          color: '#d9d9d9',
          borderRadius: [2, 2, 0, 0]
        },
        barWidth: '60%'
      }
    ]
  }

  return (
    <Card bordered={false}>
      {/* 标题栏 */}
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 16 }}>
        <div style={{ display: 'flex', alignItems: 'center', gap: 16 }}>
          <span style={{ fontSize: 16, fontWeight: 500 }}>告警趋势</span>
          <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
            <span style={{ fontSize: 14, color: '#666' }}>汇聚周期</span>
            <Select
              value={timeRange}
              onChange={setTimeRange}
              style={{ width: 100 }}
              size="small"
            >
              <Select.Option value="auto">Auto</Select.Option>
              <Select.Option value="1h">1小时</Select.Option>
              <Select.Option value="1d">1天</Select.Option>
              <Select.Option value="1w">1周</Select.Option>
            </Select>
          </div>
        </div>
        <Button
          icon={<CameraOutlined />}
          size="small"
          type="text"
        />
      </div>

      {/* 图表 */}
      <div style={{ height: 400 }}>
        <ReactECharts option={option} style={{ height: '100%', width: '100%' }} />
      </div>
    </Card>
  )
}
