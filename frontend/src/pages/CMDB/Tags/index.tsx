import { useEffect, useState } from 'react'
import { Table, Button, Tag, Space, Modal, Form, Input, InputNumber, Select, message, Popconfirm, Card, Row, Col, Statistic, Tabs } from 'antd'
import { PlusOutlined, EditOutlined, DeleteOutlined, ReloadOutlined } from '@ant-design/icons'
import { Tags as TagsIcon, Tag as TagIcon } from 'lucide-react'
import type { ColumnsType } from 'antd/es/table'
import { useTagStore, TagCategory, Tag as TagType, TagStat } from '@/stores/tagStore'

export default function TagsPage() {
  const {
    categories,
    tags,
    stats,
    loading,
    fetchCategories,
    fetchTags,
    fetchStats,
    createCategory,
    updateCategory,
    deleteCategory,
    createTag,
    updateTag,
    deleteTag,
  } = useTagStore()

  const [activeTab, setActiveTab] = useState<'list' | 'stats'>('list')
  const [selectedCategory, setSelectedCategory] = useState<number | null>(null)
  const [isModalOpen, setIsModalOpen] = useState(false)
  const [modalType, setModalType] = useState<'category' | 'tag'>('category')
  const [editingItem, setEditingItem] = useState<TagCategory | TagType | null>(null)
  const [form] = Form.useForm()

  useEffect(() => {
    fetchCategories()
    fetchTags()
    fetchStats()
  }, [fetchCategories, fetchTags, fetchStats])

  // 标签分类表格列
  const categoryColumns: ColumnsType<TagCategory> = [
    { title: 'ID', dataIndex: 'id', width: 80, key: 'id' },
    {
      title: '分类名称',
      dataIndex: 'name',
      key: 'name',
      render: (text: string) => <code>{text}</code>,
    },
    { title: '显示名称', dataIndex: 'display_name', key: 'display_name' },
    { title: '描述', dataIndex: 'description', key: 'description', render: (text) => text || '-' },
    { title: '标签数', dataIndex: 'id', key: 'count', render: (id: number) => {
      const count = tags.filter(t => t.category_id === id).length
      return <span>{count}</span>
    }},
    {
      title: '系统预设',
      dataIndex: 'is_system',
      width: 100,
      key: 'is_system',
      render: (isSystem: boolean) => (
        <Tag color={isSystem ? 'blue' : 'default'}>{isSystem ? '是' : '否'}</Tag>
      ),
    },
    {
      title: '操作',
      key: 'action',
      width: 150,
      render: (_: any, record: TagCategory) => (
        <Space size="small">
          <Button
            type="link"
            size="small"
            icon={<EditOutlined size={14} />}
            onClick={() => handleEditCategory(record)}
            disabled={record.is_system}
          >
            编辑
          </Button>
          <Popconfirm
            title="确认删除"
            description="删除后无法恢复"
            onConfirm={() => handleDeleteCategory(record.id)}
            okText="确定"
            cancelText="取消"
            disabled={record.is_system}
          >
            <Button type="link" size="small" danger icon={<DeleteOutlined size={14} />} disabled={record.is_system}>
              删除
            </Button>
          </Popconfirm>
        </Space>
      ),
    },
  ]

  // 标签表格列
  const tagColumns: ColumnsType<TagType> = [
    { title: 'ID', dataIndex: 'id', width: 80, key: 'id' },
    {
      title: '标签',
      dataIndex: 'color',
      key: 'color',
      width: 60,
      render: (color: string) => (
        <span
          style={{
            display: 'inline-block',
            width: 20,
            height: 20,
            backgroundColor: color || '#ccc',
            borderRadius: 4,
          }}
        />
      ),
    },
    { title: '名称', dataIndex: 'name', key: 'name', render: (text: string) => <code>{text}</code> },
    { title: '显示名称', dataIndex: 'display_name', key: 'display_name' },
    {
      title: '分类',
      dataIndex: 'category_id',
      key: 'category_id',
      render: (categoryId: number) => {
        const category = categories.find(c => c.id === categoryId)
        return category ? (
          <Tag color={category.color || 'default'}>{category.display_name}</Tag>
        ) : '-'
      },
    },
    { title: '描述', dataIndex: 'description', key: 'description', render: (text) => text || '-' },
    {
      title: '使用次数',
      dataIndex: 'usage_count',
      width: 100,
      key: 'usage_count',
      render: (count: number) => <Tag color="blue">{count}</Tag>,
    },
    {
      title: '操作',
      key: 'action',
      width: 150,
      render: (_: any, record: TagType) => (
        <Space size="small">
          <Button type="link" size="small" icon={<EditOutlined />} onClick={() => handleEditTag(record)}>
            编辑
          </Button>
          <Popconfirm
            title="确认删除"
            description="删除后无法恢复"
            onConfirm={() => handleDeleteTag(record.id)}
            okText="确定"
            cancelText="取消"
          >
            <Button type="link" size="small" danger icon={<DeleteOutlined />}>
              删除
            </Button>
          </Popconfirm>
        </Space>
      ),
    },
  ]

  // 统计表格列
  const statsColumns: ColumnsType<TagStat> = [
    { title: '标签', dataIndex: 'display_name', key: 'display_name' },
    {
      title: '分类',
      dataIndex: 'category_name',
      key: 'category_name',
      render: (text: string) => <Tag>{text}</Tag>,
    },
    {
      title: '颜色',
      dataIndex: 'color',
      key: 'color',
      width: 80,
      render: (color: string) => (
        <span
          style={{
            display: 'inline-block',
            width: 20,
            height: 20,
            backgroundColor: color || '#ccc',
            borderRadius: 4,
          }}
        />
      ),
    },
    {
      title: '使用次数',
      dataIndex: 'usage_count',
      key: 'usage_count',
      width: 100,
      render: (count: number) => <Statistic value={count} valueStyle={{ fontSize: 14 }} />,
    },
  ]

  const handleCreateCategory = () => {
    setModalType('category')
    setEditingItem(null)
    form.resetFields()
    setIsModalOpen(true)
  }

  const handleCreateTag = () => {
    setModalType('tag')
    setEditingItem(null)
    form.resetFields()
    setIsModalOpen(true)
  }

  const handleEditCategory = (record: TagCategory) => {
    setModalType('category')
    setEditingItem(record)
    form.setFieldsValue({
      name: record.name,
      display_name: record.display_name,
      description: record.description,
      color: record.color,
      icon: record.icon,
      sort_order: record.sort_order,
    })
    setIsModalOpen(true)
  }

  const handleEditTag = (record: TagType) => {
    setModalType('tag')
    setEditingItem(record)
    form.setFieldsValue({
      category_id: record.category_id,
      name: record.name,
      display_name: record.display_name,
      description: record.description,
      color: record.color,
    })
    setIsModalOpen(true)
  }

  const handleDeleteCategory = async (id: number) => {
    try {
      await deleteCategory(id)
      message.success('删除成功')
    } catch (error) {
      message.error('删除失败')
    }
  }

  const handleDeleteTag = async (id: number) => {
    try {
      await deleteTag(id)
      message.success('删除成功')
    } catch (error) {
      message.error('删除失败')
    }
  }

  const handleSubmit = async () => {
    try {
      const values = await form.validateFields()

      if (modalType === 'category') {
        if (editingItem) {
          await updateCategory((editingItem as TagCategory).id, values)
          message.success('更新成功')
        } else {
          await createCategory(values)
          message.success('创建成功')
        }
      } else {
        if (editingItem) {
          await updateTag((editingItem as TagType).id, values)
          message.success('更新成功')
        } else {
          await createTag(values)
          message.success('创建成功')
        }
      }

      setIsModalOpen(false)
      form.resetFields()
    } catch (error) {
      message.error('操作失败')
    }
  }

  return (
    <div className="p-8">
      {/* 页面头部 */}
      <div className="mb-6">
        <h1 className="text-2xl font-semibold text-gray-900 dark:text-text-primary mb-2">标签管理</h1>
        <p className="text-gray-600 dark:text-text-secondary">管理标签分类和标签定义</p>
      </div>

      {/* 工具栏 */}
      <div className="mb-6 flex gap-4 items-center bg-white dark:bg-bg-secondary p-4 rounded-lg border border-gray-200 dark:border-white/8">
        <Select
          placeholder="分类筛选"
          value={selectedCategory}
          onChange={(value) => {
            setSelectedCategory(value)
            fetchTags(value)
          }}
          allowClear
          style={{ width: 200 }}
        >
          {categories.map(cat => (
            <Select.Option key={cat.id} value={cat.id}>
              {cat.display_name} ({cat.tags?.length || 0})
            </Select.Option>
          ))}
        </Select>
        <Button icon={<ReloadOutlined />} onClick={() => { fetchCategories(); fetchTags(); fetchStats() }}>
          刷新
        </Button>
        <div className="flex-1" />
        <Button icon={<PlusOutlined />} onClick={handleCreateCategory} style={{ marginRight: 8 }}>
          添加分类
        </Button>
        <Button type="primary" icon={<PlusOutlined />} onClick={handleCreateTag}>
          添加标签
        </Button>
      </div>

      {/* 标签统计卡片 */}
      <Row gutter={16} className="mb-6">
        <Col span={6}>
          <Card>
            <Statistic
              title="标签分类"
              value={categories.length}
              prefix={<TagsIcon size={16} />}
              valueStyle={{ color: '#3f8600' }}
            />
          </Card>
        </Col>
        <Col span={6}>
          <Card>
            <Statistic
              title="标签总数"
              value={tags.length}
              prefix={<TagIcon size={16} />}
              valueStyle={{ color: '#cf1322' }}
            />
          </Card>
        </Col>
        <Col span={6}>
          <Card>
            <Statistic
              title="系统预设"
              value={categories.filter(c => c.is_system).length}
              valueStyle={{ color: '#1890ff' }}
            />
          </Card>
        </Col>
        <Col span={6}>
          <Card>
            <Statistic
              title="自定义"
              value={categories.filter(c => !c.is_system).length}
              valueStyle={{ color: '#722ed1' }}
            />
          </Card>
        </Col>
      </Row>

      {/* 内容区域 */}
      <Tabs
        activeKey={activeTab}
        onChange={(key) => setActiveTab(key as 'list' | 'stats')}
        items={[
          {
            key: 'list',
            label: '标签列表',
            children: (
              <>
                {/* 标签分类 */}
                <div className="mb-6">
                  <h3 className="text-lg font-semibold mb-4">标签分类</h3>
                  <Table
                    columns={categoryColumns}
                    dataSource={categories}
                    rowKey="id"
                    pagination={false}
                    size="small"
                    className="bg-white dark:bg-bg-secondary"
                  />
                </div>

                {/* 标签列表 */}
                <div>
                  <h3 className="text-lg font-semibold mb-4">标签列表</h3>
                  <Table
                    columns={tagColumns}
                    dataSource={selectedCategory
                      ? tags.filter(t => t.category_id === selectedCategory)
                      : tags
                    }
                    rowKey="id"
                    loading={loading}
                    pagination={false}
                    className="bg-white dark:bg-bg-secondary"
                  />
                </div>
              </>
            ),
          },
          {
            key: 'stats',
            label: '使用统计',
            children: (
              <Table
                columns={statsColumns}
                dataSource={stats}
                rowKey="id"
                pagination={false}
                className="bg-white dark:bg-bg-secondary"
              />
            ),
          },
        ]}
      />

      {/* 创建/编辑弹窗 */}
      <Modal
        title={modalType === 'category' ? (editingItem ? '编辑分类' : '添加分类') : (editingItem ? '编辑标签' : '添加标签')}
        open={isModalOpen}
        onCancel={() => {
          setIsModalOpen(false)
          form.resetFields()
        }}
        onOk={handleSubmit}
        width={600}
      >
        <Form form={form} layout="vertical" initialValues={{ sort_order: 0 }}>
          {modalType === 'category' && (
            <>
              <Form.Item
                label="分类名称"
                name="name"
                rules={[{ required: true, message: '请输入分类名称' }]}
              >
                <Input placeholder="例如: environment" />
              </Form.Item>

              <Form.Item
                label="显示名称"
                name="display_name"
                rules={[{ required: true, message: '请输入显示名称' }]}
              >
                <Input placeholder="例如: 环境" />
              </Form.Item>

              <Form.Item label="描述" name="description">
                <Input.TextArea placeholder="分类描述" rows={2} />
              </Form.Item>

              <Form.Item label="颜色" name="color" rules={[{ required: true, message: '请选择颜色' }]}>
                <Input type="color" className="w-32" />
              </Form.Item>

              <Form.Item label="图标" name="icon">
                <Select placeholder="选择图标">
                  <Select.Option value="environment">环境</Select.Option>
                  <Select.Option value="business">业务</Select.Option>
                  <Select.Option value="location">位置</Select.Option>
                  <Select.Option value="tag">标签</Select.Option>
                </Select>
              </Form.Item>

              <Form.Item label="排序" name="sort_order">
                <InputNumber min={0} style={{ width: '100%' }} />
              </Form.Item>
            </>
          )}

          {modalType === 'tag' && (
            <>
              <Form.Item label="所属分类" name="category_id">
                <Select placeholder="选择分类">
                  {categories.map(cat => (
                    <Select.Option key={cat.id} value={cat.id}>
                      {cat.display_name}
                    </Select.Option>
                  ))}
                </Select>
              </Form.Item>

              <Form.Item
                label="标签名称"
                name="name"
                rules={[{ required: true, message: '请输入标签名称' }]}
              >
                <Input placeholder="例如: prod" />
              </Form.Item>

              <Form.Item
                label="显示名称"
                name="display_name"
                rules={[{ required: true, message: '请输入显示名称' }]}
              >
                <Input placeholder="例如: 生产环境" />
              </Form.Item>

              <Form.Item label="描述" name="description">
                <Input.TextArea placeholder="标签描述" rows={2} />
              </Form.Item>

              <Form.Item label="颜色" name="color" rules={[{ required: true, message: '请选择颜色' }]}>
                <Input type="color" className="w-32" />
              </Form.Item>
            </>
          )}
        </Form>
      </Modal>
    </div>
  )
}
