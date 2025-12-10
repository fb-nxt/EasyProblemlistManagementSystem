import React, { useState, useEffect } from 'react'
import {
  Modal,
  Form,
  Input,
  Select,
  Button,
  Space,
  InputNumber,
  Radio,
  message,
  Tabs
} from 'antd'
import axios from 'axios'
import AIGeneratePanel from './AIGeneratePanel'

const { Option } = Select
const { TextArea } = Input
const { TabPane } = Tabs

const CreateQuestionModal = ({ visible, onClose, mode, question }) => {
  const [form] = Form.useForm()
  const [activeTab, setActiveTab] = useState('manual')
  const [loading, setLoading] = useState(false)
  const [aiGeneratedQuestions, setAiGeneratedQuestions] = useState([])
  const [selectedAiQuestions, setSelectedAiQuestions] = useState([])

  useEffect(() => {
    if (mode === 'edit' && question) {
      form.setFieldsValue({
        type: question.type,
        content: question.content,
        difficulty: question.difficulty,
        language: question.language,
        answer: question.answer
      })
    } else {
      form.resetFields()
    }
  }, [mode, question, form])

  const handleSubmit = async (values) => {
    setLoading(true)
    try {
      if (mode === 'edit') {
        await axios.put(`/api/questions/${question.id}`, values)
        message.success('更新成功')
      } else {
        await axios.post('/api/questions', values)
        message.success('创建成功')
      }
      onClose()
    } catch (error) {
      message.error(mode === 'edit' ? '更新失败' : '创建失败')
    } finally {
      setLoading(false)
    }
  }

  const handleAiQuestionsSelect = (selectedQuestions) => {
    setSelectedAiQuestions(selectedQuestions)
  }

  const handleAddAiQuestions = async () => {
    if (selectedAiQuestions.length === 0) {
      message.warning('请先选择要添加的题目')
      return
    }

    setLoading(true)
    try {
      // 批量添加AI生成的题目
      const promises = selectedAiQuestions.map(q =>
        axios.post('/api/questions', q)
      )
      await Promise.all(promises)
      message.success(`成功添加 ${selectedAiQuestions.length} 道题目`)
      onClose()
    } catch (error) {
      message.error('添加题目失败')
    } finally {
      setLoading(false)
    }
  }

  return (
    <Modal
      title={mode === 'edit' ? '编辑题目' : '出题'}
      open={visible}
      onCancel={onClose}
      width={800}
      footer={null}
      destroyOnClose
    >
      <Tabs activeKey={activeTab} onChange={setActiveTab}>
        <TabPane tab="手工出题" key="manual">
          <Form
            form={form}
            layout="vertical"
            onFinish={handleSubmit}
            initialValues={{
              type: 'single_choice',
              difficulty: 'medium'
            }}
          >
            <Form.Item
              name="type"
              label="题目类型"
              rules={[{ required: true, message: '请选择题目类型' }]}
            >
              <Select>
                <Option value="single_choice">单选题</Option>
                <Option value="multiple_choice">多选题</Option>
                <Option value="programming">编程题</Option>
              </Select>
            </Form.Item>

            <Form.Item
              name="content"
              label="题目内容"
              rules={[{ required: true, message: '请输入题目内容' }]}
            >
              <TextArea rows={4} placeholder="请输入题目内容" />
            </Form.Item>

            <Form.Item
              name="difficulty"
              label="难度"
              rules={[{ required: true, message: '请选择难度' }]}
            >
              <Select>
                <Option value="easy">简单</Option>
                <Option value="medium">中等</Option>
                <Option value="hard">困难</Option>
              </Select>
            </Form.Item>

            <Form.Item
              name="language"
              label="编程语言"
              shouldUpdate={(prevValues, curValues) => 
                prevValues.type !== curValues.type
              }
            >
              {({ getFieldValue }) =>
                getFieldValue('type') === 'programming' ? (
                  <Select placeholder="请选择编程语言">
                    <Option value="Go">Go</Option>
                    <Option value="JavaScript">JavaScript</Option>
                    <Option value="Python">Python</Option>
                    <Option value="Java">Java</Option>
                  </Select>
                ) : null
              }
            </Form.Item>

            <Form.Item
              name="answer"
              label="答案"
              shouldUpdate={(prevValues, curValues) => 
                prevValues.type !== curValues.type
              }
            >
              {({ getFieldValue }) => {
                const type = getFieldValue('type')
                if (type === 'single_choice' || type === 'multiple_choice') {
                  return (
                    <Input placeholder="请输入正确答案（如：A 或 A,B,C）" />
                  )
                }
                return null
              }}
            </Form.Item>

            <Form.Item>
              <Space style={{ float: 'right' }}>
                <Button onClick={onClose}>取消</Button>
                <Button type="primary" htmlType="submit" loading={loading}>
                  {mode === 'edit' ? '更新' : '创建'}
                </Button>
              </Space>
            </Form.Item>
          </Form>
        </TabPane>

        <TabPane tab="AI出题" key="ai">
          <AIGeneratePanel onQuestionsGenerated={setAiGeneratedQuestions} />
          
          {aiGeneratedQuestions.length > 0 && (
            <div style={{ marginTop: 24 }}>
              <Space style={{ float: 'right', marginBottom: 16 }}>
                <Button
                  type="primary"
                  onClick={handleAddAiQuestions}
                  loading={loading}
                >
                  添加选中题目到题库
                </Button>
              </Space>
            </div>
          )}
        </TabPane>
      </Tabs>
    </Modal>
  )
}

export default CreateQuestionModal