import { Card, Table, Row, Col, Statistic, Descriptions } from 'antd'
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
}

export default function ServerHardwareInfo({ attributes }: HardwareInfo) {
  const {
    hostname,
    system_serial,
    last_hardware_report,
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
    {
      label: '最后上报时间',
      value: last_hardware_report ? new Date(last_hardware_report).toLocaleString('zh-CN') : '-',
    },
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
              title="内存插槽"
              value={memory_slots || 0}
              suffix="条"
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
        </Descriptions>
      </Card>

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
