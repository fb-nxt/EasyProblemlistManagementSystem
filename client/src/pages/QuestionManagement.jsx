import React, { useState, useEffect } from 'react'
import {
  Table,
  Button,
  Space,
  Input,
  Select,
  Popconfirm,
  message,
  Card,
  Row,
  Col,
  Tag
} from 'antd'
import {
  SearchOutlined,
  EditOutlined,
  DeleteOutlined,
  SyncOutlined
} from '@ant-design/icons'
import axios from 'axios'
import CreateQuestionModal from '../components/CreateQuestionModal'

const { Search } = Input
const { Option } = Select

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

  const columns = [
    {
      title: 'ID',
      dataIndex: 'id',
      key: 'id',
      width: 80
    },
    {
      title: '题目类型',
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
      title: '题目内容',
      dataIndex: 'content',
      key: 'content',
      ellipsis: true
    },
    {
      title: '难度',
      dataIndex: 'difficulty',
      key: 'difficulty',
      width: 100,
      render: (difficulty) => {
        const difficultyMap = {
          easy: { text: '简单', color: 'green' },
          medium: { text: '中等', color: 'orange' },
          hard: { text: '困难', color: 'red' }
        }
        const config = difficultyMap[difficulty] || { text: difficulty, color: 'default' }
        return <Tag color={config.color}>{config.text}</Tag>
      }
    },
    {
      title: '编程语言',
      dataIndex: 'language',
      key: 'language',
      width: 120
    },
    {
      title: '创建时间',
      dataIndex: 'created_at',
      key: 'created_at',
      width: 180,
      render: (date) => new Date(date).toLocaleString()
    },
    {
      title: '操作',
      key: 'action',
      width: 150,
      render: (_, record) => (
        <Space size="middle">
          <Button
            type="link"
            icon={<EditOutlined />}
            onClick={() => handleEdit(record)}
          >
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
  ]

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
      setPagination({
        ...pagination,
        total: response.data.total
      })
    } catch (error) {
      message.error('获取题目列表失败')
      console.error('Error fetching questions:', error)
    } finally {
      setLoading(false)
    }
  }

  const handleTableChange = (pagination) => {
    setPagination(pagination)
  }

  const handleSearch = (value) => {
    setFilters({ ...filters, keyword: value })
    setPagination({ ...pagination, current: 1 })
  }

  const handleTypeFilter = (value) => {
    setFilters({ ...filters, type: value })
    setPagination({ ...pagination, current: 1 })
  }

  const handleDifficultyFilter = (value) => {
    setFilters({ ...filters, difficulty: value })
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

  const rowSelection = {
    selectedRowKeys,
    onChange: setSelectedRowKeys
  }

  return (
    <Card
      title="题库管理"
      extra={
        <Space>
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
      }
    >
      <Row gutter={[16, 16]} style={{ marginBottom: 16 }}>
        <Col span={6}>
          <Search
            placeholder="搜索题目内容"
            onSearch={handleSearch}
            enterButton={<SearchOutlined />}
            allowClear
          />
        </Col>
        <Col span={4}>
          <Select
            placeholder="题目类型"
            style={{ width: '100%' }}
            onChange={handleTypeFilter}
            allowClear
          >
            <Option value="single_choice">单选题</Option>
            <Option value="multiple_choice">多选题</Option>
            <Option value="programming">编程题</Option>
          </Select>
        </Col>
        <Col span={4}>
          <Select
            placeholder="难度"
            style={{ width: '100%' }}
            onChange={handleDifficultyFilter}
            allowClear
          >
            <Option value="easy">简单</Option>
            <Option value="medium">中等</Option>
            <Option value="hard">困难</Option>
          </Select>
        </Col>
      </Row>

      <Table
        columns={columns}
        dataSource={questions}
        rowKey="id"
        loading={loading}
        pagination={pagination}
        onChange={handleTableChange}
        rowSelection={rowSelection}
      />

      {isModalVisible && (
        <CreateQuestionModal
          visible={isModalVisible}
          onClose={handleModalClose}
          mode={editingQuestion ? 'edit' : 'create'}
          question={editingQuestion}
        />
      )}
    </Card>
  )
}

export default QuestionManagement