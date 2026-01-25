import { Card, Table, Row, Col, Statistic, Descriptions, Tag } from 'antd'
import {
  HddOutlined,
  ThunderboltOutlined,
  WifiOutlined,
  EyeOutlined,
  DashboardOutlined,
  ApiOutlined,
} from '@ant-design/icons'

interface HardwareInfo {
  attributes: Record<string, any>
  roles?: any[]
  tags?: any[]
}

// 格式化内存大小（将 KB 转换为合适的单位）
function formatMemorySize(memoryStr: string): string {
  if (!memoryStr) return '-'

  // 解析字符串，例如 "2113464020 kB"
  const match = memoryStr.match(/^([\d.]+)\s*(\w+)$/)
  if (!match) return memoryStr

  const value = parseFloat(match[1])
  const unit = match[2].toUpperCase()

  // 如果已经是 kB，转换为 GB
  if (unit === 'KB' || unit === 'K') {
    const gb = value / 1024 / 1024
    if (gb >= 1024) {
      return `${(gb / 1024).toFixed(2)} TB`
    }
    return `${gb.toFixed(2)} GB`
  }

  // 如果是 MB，转换为 GB
  if (unit === 'MB' || unit === 'M') {
    const gb = value / 1024
    return `${gb.toFixed(2)} GB`
  }

  return memoryStr
}

export default function ServerHardwareInfo({ attributes, roles = [], tags = [] }: HardwareInfo) {
  const {
    hostname,
    system_serial,
    last_hardware_report,
    data_center,
    rack_position,
    cpu_model,
    cpu_cores,
    cpu_threads,
    cpu_sockets,
    os_name,
    os_version,
    kernel_version,
    architecture,
    memory_total,
    memory_info,
    memory_slots,
    gpu_info,
    gpu_count,
    network_info,
    network_count,
    storage_info,
    storage_count,
    total_storage_gb,
    power_supply_info,
    total_power_capacity_w,
    optical_modules_info,
  } = attributes

  // 基本信息
  const basicInfo = [
    { label: '主机名', value: hostname || '-' },
    { label: '系统序列号', value: system_serial || '-' },
    { label: '机房', value: data_center || '-' },
    { label: '机架位', value: rack_position || '-' },
    {
      label: '最后上报时间',
      value: last_hardware_report ? new Date(last_hardware_report).toLocaleString('zh-CN') : '-',
    },
  ]

  // 系统信息
  const systemInfo = [
    { label: '操作系统', value: os_name || '-' },
    { label: '系统版本', value: os_version || '-' },
    { label: '内核版本', value: kernel_version || '-' },
    { label: '系统架构', value: architecture || '-' },
  ]

  // CPU信息
  const cpuInfo = [
    { label: 'CPU型号', value: cpu_model || '-' },
    { label: '物理CPU数量', value: cpu_sockets || '-' },
    { label: '物理核心数', value: cpu_cores || '-' },
    { label: '逻辑线程数', value: cpu_threads || '-' },
  ]

  // 内存表格列
  const memoryColumns = [
    { title: '插槽', dataIndex: 'locator', key: 'locator', width: 120 },
    { title: '类型', dataIndex: 'type', key: 'type', width: 80 },
    { title: '速度', dataIndex: 'speed', key: 'speed', width: 120 },
    { title: '制造商', dataIndex: 'manufacturer', key: 'manufacturer', width: 120 },
    { title: '型号', dataIndex: 'part_number', key: 'part_number', width: 180 },
    { title: '序列号', dataIndex: 'serial', key: 'serial', width: 240 },
  ]

  // GPU表格列
  const gpuColumns = [
    { title: 'GPU ID', dataIndex: 'gpu_id', key: 'gpu_id', width: 100 },
    { title: '产品名称', dataIndex: 'product_name', key: 'product_name' },
    { title: '序列号', dataIndex: 'serial', key: 'serial', width: 200 },
  ]

  // 网卡表格列
  const networkColumns = [
    { title: 'MAC地址', dataIndex: 'mac', key: 'mac', width: 180 },
    { title: '厂商', dataIndex: 'vendor', key: 'vendor', width: 180 },
    { title: '型号', dataIndex: 'model', key: 'model' },
    { title: '速度', dataIndex: 'speed', key: 'speed', width: 100 },
    { title: '固件', dataIndex: 'firmware', key: 'firmware', width: 200 },
    { title: '序列号', dataIndex: 'serial', key: 'serial', width: 150 },
  ]

  // 存储表格列
  const storageColumns = [
    { title: '设备名', dataIndex: 'name', key: 'name', width: 120 },
    { title: '容量', dataIndex: 'size', key: 'size', width: 120 },
    { title: '型号', dataIndex: 'model', key: 'model' },
    { title: '序列号', dataIndex: 'serial', key: 'serial', width: 200 },
  ]

  // 电源表格列
  const powerSupplyColumns = [
    { title: '位置', dataIndex: 'location', key: 'location', width: 100 },
    { title: '名称', dataIndex: 'name', key: 'name', width: 150 },
    { title: '制造商', dataIndex: 'manufacturer', key: 'manufacturer', width: 150 },
    { title: '型号', dataIndex: 'model', key: 'model', width: 150 },
    { title: '容量', dataIndex: 'capacity', key: 'capacity', width: 120 },
    { title: '序列号', dataIndex: 'serial', key: 'serial', width: 180 },
  ]

  // 光模块表格列
  const opticalColumns = [
    { title: 'PCI地址', dataIndex: 'pci_addr', key: 'pci_addr', width: 150 },
    { title: 'IB/CA', dataIndex: 'ib_ca', key: 'ib_ca', width: 150 },
    { title: '端口', dataIndex: 'port', key: 'port', width: 250 },
    { title: '类型', dataIndex: 'identifier', key: 'identifier', width: 120 },
    { title: '厂商', dataIndex: 'vendor_name', key: 'vendor_name', width: 120 },
    { title: '型号', dataIndex: 'part_number', key: 'part_number', width: 150 },
    { title: '速度', dataIndex: 'speed', key: 'speed', width: 100 },
    { title: '温度', dataIndex: 'temperature', key: 'temperature', width: 80 },
    { title: '波长', dataIndex: 'wavelength', key: 'wavelength', width: 100 },
    { title: '序列号', dataIndex: 'serial_number', key: 'serial_number', width: 180 },
  ]

  return (
    <div className="space-y-4">
      {/* 硬件概览统计 */}
      <Card title={<><DashboardOutlined /> 硬件概览</>}>
        <Row gutter={16}>
          <Col span={4}>
            <Statistic
              title="物理CPU"
              value={cpu_sockets || 0}
              suffix="颗"
              prefix={<ApiOutlined />}
              valueStyle={{ color: '#1890ff' }}
            />
          </Col>
          <Col span={4}>
            <Statistic
              title="物理核心"
              value={cpu_cores || 0}
              suffix="核"
              prefix={<ApiOutlined />}
              valueStyle={{ color: '#1890ff' }}
            />
          </Col>
          <Col span={4}>
            <Statistic
              title="逻辑线程"
              value={cpu_threads || 0}
              suffix="线程"
              prefix={<ApiOutlined />}
              valueStyle={{ color: '#1890ff' }}
            />
          </Col>
          <Col span={3}>
            <Statistic
              title="内存插槽"
              value={memory_slots || 0}
              suffix="条"
              prefix={<ApiOutlined />}
              valueStyle={{ color: '#3f8600' }}
            />
          </Col>
          <Col span={3}>
            <Statistic
              title="内存总量"
              value={formatMemorySize(memory_total || '').split(' ')[0]}
              suffix={formatMemorySize(memory_total || '').split(' ')[1] || ''}
              prefix={<ApiOutlined />}
              valueStyle={{ color: '#3f8600' }}
            />
          </Col>
          <Col span={4}>
            <Statistic
              title="GPU数量"
              value={gpu_count || 0}
              suffix="个"
              prefix={<EyeOutlined />}
              valueStyle={{ color: '#cf1322' }}
            />
          </Col>
          <Col span={4}>
            <Statistic
              title="网卡数量"
              value={network_count || 0}
              suffix="个"
              prefix={<WifiOutlined />}
              valueStyle={{ color: '#1890ff' }}
            />
          </Col>
          <Col span={4}>
            <Statistic
              title="存储设备"
              value={storage_count || 0}
              suffix="个"
              prefix={<HddOutlined />}
              valueStyle={{ color: '#722ed1' }}
            />
          </Col>
          <Col span={4}>
            <Statistic
              title="总存储容量"
              value={total_storage_gb?.toFixed(1) || 0}
              suffix="GB"
              prefix={<HddOutlined />}
              valueStyle={{ color: '#722ed1' }}
            />
          </Col>
          <Col span={4}>
            <Statistic
              title="电源容量"
              value={total_power_capacity_w || 0}
              suffix="W"
              prefix={<ThunderboltOutlined />}
              valueStyle={{ color: '#fa8c16' }}
            />
          </Col>
        </Row>
      </Card>

      {/* 基本信息 */}
      <Card title="基本信息">
        <Descriptions column={3} bordered size="small">
          {basicInfo.map((info, index) => (
            <Descriptions.Item key={index} label={info.label}>
              {info.value}
            </Descriptions.Item>
          ))}
          <Descriptions.Item label="角色" span={3}>
            {roles && roles.length > 0 ? (
              <div className="flex flex-wrap gap-2">
                {roles.map((role: any) => (
                  <Tag key={role.id} color="blue">
                    {role.display_name || role.name}
                  </Tag>
                ))}
              </div>
            ) : (
              '-'
            )}
          </Descriptions.Item>
          <Descriptions.Item label="标签" span={3}>
            {tags && tags.length > 0 ? (
              <div className="flex flex-wrap gap-2">
                {tags.map((tag: any) => (
                  <Tag key={tag.id} color={tag.color || 'default'}>
                    {tag.display_name || tag.name}
                  </Tag>
                ))}
              </div>
            ) : (
              '-'
            )}
          </Descriptions.Item>
        </Descriptions>
      </Card>

      {/* CPU信息 */}
      {cpu_model && (
        <Card title={<><ApiOutlined /> CPU信息</>}>
          <Descriptions column={2} bordered size="small">
            {cpuInfo.map((info, index) => (
              <Descriptions.Item key={index} label={info.label}>
                {info.value}
              </Descriptions.Item>
            ))}
          </Descriptions>
        </Card>
      )}

      {/* 系统信息 */}
      {(os_name || kernel_version) && (
        <Card title={<><DashboardOutlined /> 系统信息</>}>
          <Descriptions column={2} bordered size="small">
            {systemInfo.map((info, index) => (
              <Descriptions.Item key={index} label={info.label}>
                {info.value}
              </Descriptions.Item>
            ))}
          </Descriptions>
        </Card>
      )}

      {/* 内存信息 */}
      {Array.isArray(memory_info) && memory_info.length > 0 && (
        <Card title={<><ApiOutlined /> 内存信息 ({memory_info.length}条)</>}>
          <Table
            columns={memoryColumns}
            dataSource={memory_info}
            rowKey={(record) => record.locator + record.serial}
            size="small"
            pagination={false}
            scroll={{ x: 1000 }}
          />
        </Card>
      )}

      {/* GPU信息 */}
      {Array.isArray(gpu_info) && gpu_info.length > 0 && (
        <Card title={<><EyeOutlined /> GPU信息 ({gpu_info.length}个)</>}>
          <Table
            columns={gpuColumns}
            dataSource={gpu_info}
            rowKey={(record) => record.gpu_id}
            size="small"
            pagination={false}
          />
        </Card>
      )}

      {/* 网卡信息 */}
      {Array.isArray(network_info) && network_info.length > 0 && (
        <Card title={<><WifiOutlined /> 网卡信息 ({network_info.length}个)</>}>
          <Table
            columns={networkColumns}
            dataSource={network_info}
            rowKey={(record) => record.mac}
            size="small"
            pagination={false}
            scroll={{ x: 1400 }}
          />
        </Card>
      )}

      {/* 存储信息 */}
      {Array.isArray(storage_info) && storage_info.length > 0 && (
        <Card title={<><HddOutlined /> 存储信息 ({storage_info.length}个)</>}>
          <Table
            columns={storageColumns}
            dataSource={storage_info}
            rowKey={(record) => record.name + record.serial}
            size="small"
            pagination={false}
            scroll={{ x: 1000 }}
          />
        </Card>
      )}

      {/* 电源信息 */}
      {Array.isArray(power_supply_info) && power_supply_info.length > 0 && (
        <Card title={<><ThunderboltOutlined /> 电源信息 ({power_supply_info.length}个)</>}>
          <Table
            columns={powerSupplyColumns}
            dataSource={power_supply_info}
            rowKey={(record) => record.location}
            size="small"
            pagination={false}
            scroll={{ x: 1000 }}
          />
        </Card>
      )}

      {/* 光模块信息 */}
      {Array.isArray(optical_modules_info) && optical_modules_info.length > 0 && (
        <Card title={<><DashboardOutlined /> 光模块信息 ({optical_modules_info.length}个)</>}>
          <Table
            columns={opticalColumns}
            dataSource={optical_modules_info}
            rowKey={(record) => record.pci_addr + record.port}
            size="small"
            pagination={false}
            scroll={{ x: 1600 }}
          />
        </Card>
      )}
    </div>
  )
}
