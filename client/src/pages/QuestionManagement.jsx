import React, { useState, useEffect, useMemo } from 'react'
import {
  Table,
  Button,
  Space,
  Input,
  Popconfirm,
  message,
  Tag,
  Dropdown
} from 'antd'
import {
  SearchOutlined,
  EditOutlined,
  DeleteOutlined,
  SyncOutlined,
  PlusOutlined
} from '@ant-design/icons'
import axios from 'axios'
import CreateQuestionModal from '../components/CreateQuestionModal'

const { Search } = Input

const typeTabs = [
  { label: '全部', value: '' },
  { label: '单选题', value: 'single_choice' },
  { label: '多选题', value: 'multiple_choice' },
  { label: '编程题', value: 'programming' }
]

const QuestionManagement = () => {
  const [questions, setQuestions] = useState([])
  const [loading, setLoading] = useState(false)
  const [pagination, setPagination] = useState({
    current: 1,
    pageSize: 10,
    total: 0
  })
  const [filters, setFilters] = useState({
    type: '',
    difficulty: '',
    keyword: ''
  })
  const [selectedRowKeys, setSelectedRowKeys] = useState([])
  const [isModalVisible, setIsModalVisible] = useState(false)
  const [editingQuestion, setEditingQuestion] = useState(null)
  const [defaultTab, setDefaultTab] = useState('manual')

  const columns = useMemo(
    () => [
      {
        title: '题目',
        dataIndex: 'content',
        key: 'content',
        width : 150,
        ellipsis: true
      },
      {
        title: '题型',
        dataIndex: 'type',
        key: 'type',
        width: 120,
        render: (type) => {
          const typeMap = {
            single_choice: { text: '单选题', color: 'blue' },
            multiple_choice: { text: '多选题', color: 'green' },
            programming: { text: '编程题', color: 'orange' }
          }
          const config = typeMap[type] || { text: type, color: 'default' }
          return <Tag color={config.color}>{config.text}</Tag>
        }
      },
      {
        title: '操作',
        key: 'action',
        width: 180,
        align: 'center',
        render: (_, record) => (
          <Space size="middle">
            <Button type="link" icon={<EditOutlined />} onClick={() => handleEdit(record)}>
              编辑
            </Button>
            <Popconfirm
              title="确定删除这道题目吗？"
              onConfirm={() => handleDelete(record.id)}
              okText="确定"
              cancelText="取消"
            >
              <Button type="link" danger icon={<DeleteOutlined />}>
                删除
              </Button>
            </Popconfirm>
          </Space>
        )
      }
    ],
    []
  )

  useEffect(() => {
    fetchQuestions()
  }, [pagination.current, pagination.pageSize, filters])

  const fetchQuestions = async () => {
    setLoading(true)
    try {
      const params = {
        page: pagination.current,
        page_size: pagination.pageSize,
        ...filters
      }
      const response = await axios.get('/api/questions', { params })
      setQuestions(response.data.data)
      setPagination((prev) => ({
        ...prev,
        total: response.data.total
      }))
    } catch (error) {
      message.error('获取题目列表失败')
      console.error('Error fetching questions:', error)
    } finally {
      setLoading(false)
    }
  }

  const handleTableChange = (tablePagination) => {
    setPagination(tablePagination)
  }

  const handleSearch = (value) => {
    setFilters({ ...filters, keyword: value })
    setPagination({ ...pagination, current: 1 })
  }

  const handleTypeFilter = (value) => {
    setFilters({ ...filters, type: value })
    setPagination({ ...pagination, current: 1 })
  }

  const handleDelete = async (id) => {
    try {
      await axios.delete(`/api/questions/${id}`)
      message.success('删除成功')
      fetchQuestions()
    } catch (error) {
      message.error('删除失败')
    }
  }

  const handleBatchDelete = async () => {
    if (selectedRowKeys.length === 0) {
      message.warning('请先选择要删除的题目')
      return
    }

    try {
      await axios.delete('/api/questions', { data: selectedRowKeys })
      message.success('批量删除成功')
      setSelectedRowKeys([])
      fetchQuestions()
    } catch (error) {
      message.error('批量删除失败')
    }
  }

  const handleEdit = (question) => {
    setEditingQuestion(question)
    setIsModalVisible(true)
  }

  const handleModalClose = () => {
    setIsModalVisible(false)
    setEditingQuestion(null)
    fetchQuestions()
  }

  const handleCreateManual = () => {
    setEditingQuestion(null)
    setDefaultTab('manual')
    setIsModalVisible(true)
  }

  const handleCreateAI = () => {
    setEditingQuestion(null)
    setDefaultTab('ai')
    setIsModalVisible(true)
  }

  const addMenu = {
    onClick: ({ key }) => {
      if (key === 'ai') {
        handleCreateAI()
      } else {
        handleCreateManual()
      }
    },
    items: [
      { key: 'ai', label: 'AI出题' },
      { key: 'manual', label: '自主出题' }
    ]
  }

  const rowSelection = {
    selectedRowKeys,
    onChange: setSelectedRowKeys
  }

  return (
    <div>
      <div className="card table-card">
        <div className="page-title">题库管理</div>
        <div className="table-toolbar">
          <div className="pill-group">
            {typeTabs.map((tab) => (
              <div
                key={tab.value}
                className={`pill ${filters.type === tab.value ? 'active' : ''}`}
                onClick={() => handleTypeFilter(tab.value)}
              >
                {tab.label}
              </div>
            ))}
          </div>
          <Search
            placeholder="请输入试题名称"
            onSearch={handleSearch}
            enterButton={<SearchOutlined />}
            allowClear
            style={{ maxWidth: 240 }}
          />
          <Dropdown menu={addMenu} trigger={['click']}>
            <Button type="primary" icon={<PlusOutlined />}>
              出题
            </Button>
          </Dropdown>
          <Space style={{ marginLeft: 'auto' }}>
            <Button
              danger
              onClick={handleBatchDelete}
              disabled={selectedRowKeys.length === 0}
            >
              批量删除
            </Button>
            <Button icon={<SyncOutlined />} onClick={fetchQuestions}>
              刷新
            </Button>
          </Space>
        </div>

        <Table
          columns={columns}
          dataSource={questions}
          rowKey="id"
          loading={loading}
          pagination={pagination}
          onChange={handleTableChange}
          rowSelection={rowSelection}
        />
      </div>

      {isModalVisible && (
        <CreateQuestionModal
          visible={isModalVisible}
          onClose={handleModalClose}
          mode={editingQuestion ? 'edit' : 'create'}
          question={editingQuestion}
          defaultTab={defaultTab}
        />
      )}
    </div>
  )
}

export default QuestionManagement
