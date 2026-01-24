import { useState } from 'react'
import { Modal, Upload, Button, message, Progress, Alert } from 'antd'
import { UploadOutlined, DownloadOutlined, InboxOutlined } from '@ant-design/icons'
import type { UploadProps } from 'antd'
import { useAuthStore } from '@/stores/authStore'

interface ImportExportModalProps {
  visible: boolean
  ciTypeId: number
  ciTypeName: string
  onClose: () => void
  onSuccess: () => void
}

interface ImportResult {
  success: number
  failed: number
  errors: string[]
}

export default function ImportExportModal({
  visible,
  ciTypeId,
  ciTypeName,
  onClose,
  onSuccess,
}: ImportExportModalProps) {
  const { token } = useAuthStore()
  const [importing, setImporting] = useState(false)
  const [exporting, setExporting] = useState(false)
  const [importResult, setImportResult] = useState<ImportResult | null>(null)

  const handleExport = async () => {
    setExporting(true)
    try {
      const response = await fetch(`/api/v1/ci/export?ci_type_id=${ciTypeId}`, {
        headers: {
          Authorization: `Bearer ${token}`,
        },
      })

      if (!response.ok) {
        throw new Error('导出失败')
      }

      // 下载文件
      const blob = await response.blob()
      const url = window.URL.createObjectURL(blob)
      const a = document.createElement('a')
      a.href = url
      a.download = `${ciTypeName}_instances_${new Date().getTime()}.json`
      document.body.appendChild(a)
      a.click()
      window.URL.revokeObjectURL(url)
      document.body.removeChild(a)

      message.success('导出成功')
    } catch (error) {
      message.error('导出失败')
      console.error('Export error:', error)
    } finally {
      setExporting(false)
    }
  }

  const uploadProps: UploadProps = {
    name: 'file',
    accept: '.json',
    maxCount: 1,
    beforeUpload: () => false, // 阻止自动上传
    onChange: async (info) => {
      const file = info.file
      if (!file) return

      setImporting(true)
      setImportResult(null)

      try {
        const formData = new FormData()
        formData.append('file', file as any)

        const response = await fetch(`/api/v1/ci/import?ci_type_id=${ciTypeId}`, {
          method: 'POST',
          headers: {
            Authorization: `Bearer ${token}`,
          },
          body: formData,
        })

        const result = await response.json()

        if (result.code === 0) {
          setImportResult(result.data)
          if (result.data.failed === 0) {
            message.success(`成功导入 ${result.data.success} 条记录`)
            onSuccess()
          } else {
            message.warning(
              `导入完成：成功 ${result.data.success} 条，失败 ${result.data.failed} 条`
            )
          }
        } else {
          message.error(result.message || '导入失败')
        }
      } catch (error) {
        message.error('导入失败')
        console.error('Import error:', error)
      } finally {
        setImporting(false)
      }
    },
  }

  const handleClose = () => {
    setImportResult(null)
    onClose()
  }

  return (
    <Modal
      title={`批量导入/导出 - ${ciTypeName}`}
      open={visible}
      onCancel={handleClose}
      footer={null}
      width={600}
    >
      <div className="space-y-6">
        {/* 导出部分 */}
        <div>
          <h3 className="text-base font-medium mb-3">导出数据</h3>
          <p className="text-sm text-gray-600 mb-3">
            导出当前类型的所有CI实例为JSON格式文件
          </p>
          <Button
            type="primary"
            icon={<DownloadOutlined />}
            onClick={handleExport}
            loading={exporting}
            block
          >
            导出为JSON
          </Button>
        </div>

        {/* 分隔线 */}
        <div className="border-t border-gray-200" />

        {/* 导入部分 */}
        <div>
          <h3 className="text-base font-medium mb-3">导入数据</h3>
          <p className="text-sm text-gray-600 mb-3">
            上传JSON格式文件批量导入CI实例
          </p>

          <Upload.Dragger {...uploadProps} disabled={importing}>
            <p className="ant-upload-drag-icon">
              <InboxOutlined />
            </p>
            <p className="ant-upload-text">点击或拖拽文件到此区域上传</p>
            <p className="ant-upload-hint">支持JSON格式文件</p>
          </Upload.Dragger>

          {importing && (
            <div className="mt-4">
              <Progress percent={100} status="active" showInfo={false} />
              <p className="text-center text-sm text-gray-600 mt-2">正在导入...</p>
            </div>
          )}

          {/* 导入结果 */}
          {importResult && (
            <div className="mt-4 space-y-3">
              <Alert
                message="导入完成"
                description={
                  <div>
                    <p>成功: {importResult.success} 条</p>
                    <p>失败: {importResult.failed} 条</p>
                  </div>
                }
                type={importResult.failed === 0 ? 'success' : 'warning'}
                showIcon
              />

              {importResult.errors && importResult.errors.length > 0 && (
                <Alert
                  message="错误详情"
                  description={
                    <ul className="list-disc list-inside max-h-40 overflow-y-auto">
                      {importResult.errors.map((error, index) => (
                        <li key={index} className="text-sm">
                          {error}
                        </li>
                      ))}
                    </ul>
                  }
                  type="error"
                  showIcon
                />
              )}
            </div>
          )}
        </div>

        {/* 使用说明 */}
        <div className="bg-blue-50 dark:bg-blue-900 p-4 rounded">
          <h4 className="text-sm font-medium mb-2">数据格式说明</h4>
          <ul className="text-sm text-gray-600 dark:text-gray-300 space-y-1">
            <li>• JSON文件应包含CI实例数组</li>
            <li>• 每个实例必须包含name字段</li>
            <li>• status字段可选，默认为active</li>
            <li>• attributes字段包含CI的动态属性</li>
            <li>• 导入时会验证必填属性</li>
          </ul>
        </div>
      </div>
    </Modal>
  )
}
