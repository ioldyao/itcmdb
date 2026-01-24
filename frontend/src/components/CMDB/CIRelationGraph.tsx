import { useEffect, useState } from 'react'
import { Card, Empty, Spin, Button, Modal, Form, Select, Input, message, Space, Tag } from 'antd'
import { PlusOutlined, LinkOutlined } from '@ant-design/icons'
import { useCMDBStore, CIRelation, CIInstance } from '@/stores/cmdbStore'

interface CIRelationGraphProps {
  ciId: number
  ciName: string
}

export default function CIRelationGraph({ ciId, ciName }: CIRelationGraphProps) {
  const { fetchRelations, createRelation, fetchInstances } = useCMDBStore()
  const [relations, setRelations] = useState<CIRelation[]>([])
  const [loading, setLoading] = useState(false)
  const [modalVisible, setModalVisible] = useState(false)
  const [instances, setInstances] = useState<CIInstance[]>([])
  const [form] = Form.useForm()

  useEffect(() => {
    loadRelations()
    loadInstances()
  }, [ciId])

  const loadRelations = async () => {
    setLoading(true)
    try {
      const data = await fetchRelations(ciId)
      setRelations(data)
    } catch (error) {
      console.error('Failed to load relations:', error)
    } finally {
      setLoading(false)
    }
  }

  const loadInstances = async () => {
    try {
      await fetchInstances(0, 1, 100)
    } catch (error) {
      console.error('Failed to load instances:', error)
    }
  }

  const handleAddRelation = () => {
    form.resetFields()
    setModalVisible(true)
  }

  const handleSubmit = async () => {
    try {
      const values = await form.validateFields()
      await createRelation({
        parent_id: values.direction === 'parent' ? ciId : values.target_id,
        child_id: values.direction === 'child' ? ciId : values.target_id,
        relation_type: values.relation_type,
        description: values.description,
      })
      message.success('关系创建成功')
      setModalVisible(false)
      loadRelations()
    } catch (error) {
      message.error('创建关系失败')
    }
  }

  const getRelationTypeColor = (type: string) => {
    const colors: Record<string, string> = {
      depends_on: 'blue',
      runs_on: 'green',
      connects_to: 'orange',
      contains: 'purple',
    }
    return colors[type] || 'default'
  }

  const getRelationTypeText = (type: string) => {
    const texts: Record<string, string> = {
      depends_on: '依赖于',
      runs_on: '运行于',
      connects_to: '连接到',
      contains: '包含',
    }
    return texts[type] || type
  }

  if (loading) {
    return (
      <div className="flex justify-center items-center py-8">
        <Spin tip="加载关系图谱..." />
      </div>
    )
  }

  if (!relations || relations.length === 0) {
    return (
      <div className="py-8">
        <Empty
          image={Empty.PRESENTED_IMAGE_SIMPLE}
          description="暂无关系数据"
        >
          <Button type="primary" icon={<PlusOutlined />} onClick={handleAddRelation}>
            添加关系
          </Button>
        </Empty>

        <Modal
          title="添加CI关系"
          open={modalVisible}
          onCancel={() => setModalVisible(false)}
          onOk={handleSubmit}
          width={600}
        >
          <Form form={form} layout="vertical">
            <Form.Item
              label="关系方向"
              name="direction"
              rules={[{ required: true, message: '请选择关系方向' }]}
            >
              <Select placeholder="选择关系方向">
                <Select.Option value="parent">当前CI是父节点</Select.Option>
                <Select.Option value="child">当前CI是子节点</Select.Option>
              </Select>
            </Form.Item>

            <Form.Item
              label="目标CI"
              name="target_id"
              rules={[{ required: true, message: '请选择目标CI' }]}
            >
              <Select
                showSearch
                placeholder="选择目标CI"
                filterOption={(input, option) =>
                  (option?.label ?? '').toLowerCase().includes(input.toLowerCase())
                }
                options={instances
                  .filter((inst) => inst.id !== ciId)
                  .map((inst) => ({
                    label: `${inst.name} (${inst.ci_type?.display_name || inst.ci_type?.name})`,
                    value: inst.id,
                  }))}
              />
            </Form.Item>

            <Form.Item
              label="关系类型"
              name="relation_type"
              rules={[{ required: true, message: '请选择关系类型' }]}
            >
              <Select placeholder="选择关系类型">
                <Select.Option value="depends_on">依赖于</Select.Option>
                <Select.Option value="runs_on">运行于</Select.Option>
                <Select.Option value="connects_to">连接到</Select.Option>
                <Select.Option value="contains">包含</Select.Option>
              </Select>
            </Form.Item>

            <Form.Item label="描述" name="description">
              <Input.TextArea rows={3} placeholder="可选的关系描述" />
            </Form.Item>
          </Form>
        </Modal>
      </div>
    )
  }

  // 分组显示关系
  const parentRelations = relations.filter((r) => r.child_id === ciId)
  const childRelations = relations.filter((r) => r.parent_id === ciId)

  return (
    <div className="py-4">
      <div className="mb-4 flex justify-between items-center">
        <h3 className="text-lg font-medium">关系图谱</h3>
        <Button type="primary" icon={<PlusOutlined />} onClick={handleAddRelation}>
          添加关系
        </Button>
      </div>

      <div className="space-y-6">
        {/* 父节点关系 */}
        {parentRelations.length > 0 && (
          <Card title="依赖的CI" size="small">
            <div className="space-y-3">
              {parentRelations.map((relation) => (
                <div
                  key={relation.id}
                  className="flex items-center gap-3 p-3 bg-gray-50 dark:bg-gray-800 rounded-lg"
                >
                  <LinkOutlined className="text-blue-500" />
                  <div className="flex-1">
                    <div className="font-medium">{relation.parent?.name}</div>
                    <div className="text-sm text-gray-500">
                      {relation.parent?.ci_type?.display_name || relation.parent?.ci_type?.name}
                    </div>
                    {relation.description && (
                      <div className="text-sm text-gray-600 mt-1">{relation.description}</div>
                    )}
                  </div>
                  <Tag color={getRelationTypeColor(relation.relation_type)}>
                    {getRelationTypeText(relation.relation_type)}
                  </Tag>
                </div>
              ))}
            </div>
          </Card>
        )}

        {/* 当前节点 */}
        <Card size="small" className="bg-blue-50 dark:bg-blue-900">
          <div className="text-center">
            <div className="text-lg font-bold">{ciName}</div>
            <div className="text-sm text-gray-600 dark:text-gray-400">当前CI</div>
          </div>
        </Card>

        {/* 子节点关系 */}
        {childRelations.length > 0 && (
          <Card title="被依赖的CI" size="small">
            <div className="space-y-3">
              {childRelations.map((relation) => (
                <div
                  key={relation.id}
                  className="flex items-center gap-3 p-3 bg-gray-50 dark:bg-gray-800 rounded-lg"
                >
                  <LinkOutlined className="text-green-500" />
                  <div className="flex-1">
                    <div className="font-medium">{relation.child?.name}</div>
                    <div className="text-sm text-gray-500">
                      {relation.child?.ci_type?.display_name || relation.child?.ci_type?.name}
                    </div>
                    {relation.description && (
                      <div className="text-sm text-gray-600 mt-1">{relation.description}</div>
                    )}
                  </div>
                  <Tag color={getRelationTypeColor(relation.relation_type)}>
                    {getRelationTypeText(relation.relation_type)}
                  </Tag>
                </div>
              ))}
            </div>
          </Card>
        )}
      </div>

      {/* 添加关系模态框 */}
      <Modal
        title="添加CI关系"
        open={modalVisible}
        onCancel={() => setModalVisible(false)}
        onOk={handleSubmit}
        width={600}
      >
        <Form form={form} layout="vertical">
          <Form.Item
            label="关系方向"
            name="direction"
            rules={[{ required: true, message: '请选择关系方向' }]}
          >
            <Select placeholder="选择关系方向">
              <Select.Option value="parent">当前CI依赖于目标CI</Select.Option>
              <Select.Option value="child">目标CI依赖于当前CI</Select.Option>
            </Select>
          </Form.Item>

          <Form.Item
            label="目标CI"
            name="target_id"
            rules={[{ required: true, message: '请选择目标CI' }]}
          >
            <Select
              showSearch
              placeholder="选择目标CI"
              filterOption={(input, option) =>
                (option?.label ?? '').toLowerCase().includes(input.toLowerCase())
              }
              options={instances
                .filter((inst) => inst.id !== ciId)
                .map((inst) => ({
                  label: `${inst.name} (${inst.ci_type?.display_name || inst.ci_type?.name})`,
                  value: inst.id,
                }))}
            />
          </Form.Item>

          <Form.Item
            label="关系类型"
            name="relation_type"
            rules={[{ required: true, message: '请选择关系类型' }]}
          >
            <Select placeholder="选择关系类型">
              <Select.Option value="depends_on">依赖于</Select.Option>
              <Select.Option value="runs_on">运行于</Select.Option>
              <Select.Option value="connects_to">连接到</Select.Option>
              <Select.Option value="contains">包含</Select.Option>
            </Select>
          </Form.Item>

          <Form.Item label="描述" name="description">
            <Input.TextArea rows={3} placeholder="可选的关系描述" />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  )
}
