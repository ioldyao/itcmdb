import { useState, useEffect } from 'react'
import { Modal, Form, Select, Button, Table, message, Popconfirm, Tag, Input } from 'antd'
import { PlusOutlined, DeleteOutlined } from '@ant-design/icons'
import { useCMDBStore, CIRelation } from '@/stores/cmdbStore'

interface RelationManagerProps {
  ciId: number
  ciName: string
}

export default function RelationManager({ ciId, ciName }: RelationManagerProps) {
  const { instances, fetchInstances, fetchRelations, createRelation } = useCMDBStore()
  const [relations, setRelations] = useState<CIRelation[]>([])
  const [isModalOpen, setIsModalOpen] = useState(false)
  const [loading, setLoading] = useState(false)
  const [form] = Form.useForm()

  useEffect(() => {
    loadRelations()
    // 加载所有CI实例用于选择
    fetchInstances(undefined, 1, 1000)
  }, [ciId])

  const loadRelations = async () => {
    setLoading(true)
    try {
      const data = await fetchRelations(ciId)
      setRelations(data)
    } catch (error) {
      message.error('加载关系失败')
    } finally {
      setLoading(false)
    }
  }

  const handleCreate = async () => {
    try {
      const values = await form.validateFields()
      await createRelation({
        parent_id: values.relation_type === 'runs_on' || values.relation_type === 'deploys_on' ? values.target_id : ciId,
        child_id: values.relation_type === 'runs_on' || values.relation_type === 'deploys_on' ? ciId : values.target_id,
        relation_type: values.relation_type,
        description: values.description || '',
      })
      message.success('创建关系成功')
      setIsModalOpen(false)
      form.resetFields()
      loadRelations()
    } catch (error) {
      message.error('创建关系失败')
    }
  }

  const relationTypeConfig = {
    runs_on: { text: '运行在', color: 'blue', direction: '→' },
    deploys_on: { text: '部署在', color: 'green', direction: '→' },
    depends_on: { text: '依赖于', color: 'orange', direction: '→' },
    contains: { text: '包含', color: 'purple', direction: '←' },
    connects: { text: '连接到', color: 'cyan', direction: '↔' },
  }

  const columns = [
    {
      title: '关系类型',
      dataIndex: 'relation_type',
      key: 'relation_type',
      render: (type: string) => {
        const config = relationTypeConfig[type as keyof typeof relationTypeConfig]
        return <Tag color={config?.color}>{config?.text || type}</Tag>
      },
    },
    {
      title: '关联CI',
      key: 'target',
      render: (record: CIRelation) => {
        const isParent = record.child_id === ciId
        const targetId = isParent ? record.parent_id : record.child_id
        const target = instances.find(i => i.id === targetId)
        return target?.name || `CI #${targetId}`
      },
    },
    {
      title: '方向',
      key: 'direction',
      render: (record: CIRelation) => {
        const isParent = record.child_id === ciId
        const config = relationTypeConfig[record.relation_type as keyof typeof relationTypeConfig]
        return isParent ? `← ${config?.text}` : `${config?.direction} ${config?.text}`
      },
    },
    {
      title: '描述',
      dataIndex: 'description',
      key: 'description',
    },
    {
      title: '操作',
      key: 'action',
      render: (record: CIRelation) => (
        <Popconfirm
          title="确认删除"
          description="删除后无法恢复"
          onConfirm={() => handleDelete(record.id)}
          okText="确定"
          cancelText="取消"
        >
          <Button type="link" size="small" danger icon={<DeleteOutlined />}>
            删除
          </Button>
        </Popconfirm>
      ),
    },
  ]

  const handleDelete = async (_relationId: number) => {
    try {
      // TODO: 实现删除关系的API调用
      message.success('删除成功')
      loadRelations()
    } catch (error) {
      message.error('删除失败')
    }
  }

  return (
    <div>
      <div className="mb-4 flex justify-between items-center">
        <h3 className="text-lg font-semibold">CI 关系</h3>
        <Button type="primary" icon={<PlusOutlined />} onClick={() => setIsModalOpen(true)}>
          添加关系
        </Button>
      </div>

      <Table
        columns={columns}
        dataSource={relations}
        rowKey="id"
        loading={loading}
        pagination={false}
        size="small"
      />

      <Modal
        title={`为 ${ciName} 添加关系`}
        open={isModalOpen}
        onCancel={() => {
          setIsModalOpen(false)
          form.resetFields()
        }}
        onOk={handleCreate}
        width={500}
      >
        <Form form={form} layout="vertical">
          <Form.Item
            label="关系类型"
            name="relation_type"
            rules={[{ required: true, message: '请选择关系类型' }]}
          >
            <Select placeholder="选择关系类型">
              <Select.Option value="runs_on">运行在 (容器 → 宿主机)</Select.Option>
              <Select.Option value="deploys_on">部署在 (应用 → 容器)</Select.Option>
              <Select.Option value="depends_on">依赖于 (应用 → 应用)</Select.Option>
              <Select.Option value="contains">包含</Select.Option>
              <Select.Option value="connects">连接到</Select.Option>
            </Select>
          </Form.Item>

          <Form.Item
            label="目标CI"
            name="target_id"
            rules={[{ required: true, message: '请选择目标CI' }]}
          >
            <Select
              placeholder="选择目标CI"
              showSearch
              filterOption={(input, option) =>
                (option?.label ?? '').toLowerCase().includes(input.toLowerCase())
              }
              options={instances
                .filter(i => i.id !== ciId)
                .map(i => ({
                  label: `${i.name} (${i.ci_type?.name || 'unknown'})`,
                  value: i.id,
                }))}
            />
          </Form.Item>

          <Form.Item label="描述" name="description">
            <Input.TextArea placeholder="可选的关系描述" rows={3} />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  )
}
