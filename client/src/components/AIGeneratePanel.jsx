import React, { useState } from 'react'
import { Form, Input, Select, Button, InputNumber, Card, List, Checkbox, message } from 'antd'
import axios from 'axios'

const { Option } = Select
const { TextArea } = Input

const AIGeneratePanel = ({ onQuestionsGenerated }) => {
  const [form] = Form.useForm()
  const [loading, setLoading] = useState(false)
  const [generatedQuestions, setGeneratedQuestions] = useState([])
  const [selectedQuestions, setSelectedQuestions] = useState([])

  const handleGenerate = async (values) => {
    setLoading(true)
    try {
      const response = await axios.post('/api/ai/generate', values)
      setGeneratedQuestions(response.data.data)
      setSelectedQuestions([])
      onQuestionsGenerated(response.data.data)
      message.success('题目生成成功')
    } catch (error) {
      message.error('题目生成失败')
      console.error('Error generating questions:', error)
    } finally {
      setLoading(false)
    }
  }

  const handleQuestionSelect = (id, checked) => {
    if (checked) {
      setSelectedQuestions([...selectedQuestions, id])
    } else {
      setSelectedQuestions(selectedQuestions.filter(qId => qId !== id))
    }
  }

  return (
    <div>
      <Form
        form={form}
        layout="vertical"
        onFinish={handleGenerate}
        initialValues={{
          type: 'single_choice',
          count: 5,
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
          name="count"
          label="题目数量"
          rules={[{ required: true, message: '请输入题目数量' }]}
        >
          <InputNumber min={1} max={20} style={{ width: '100%' }} />
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

        <Form.Item name="topic" label="题目主题">
          <TextArea rows={2} placeholder="输入题目主题或关键词（可选）" />
        </Form.Item>

        <Form.Item>
          <Button type="primary" htmlType="submit" loading={loading}>
            生成题目
          </Button>
        </Form.Item>
      </Form>

      {generatedQuestions.length > 0 && (
        <Card title="生成的题目预览" style={{ marginTop: 16 }}>
          <List
            dataSource={generatedQuestions}
            renderItem={(question, index) => (
              <List.Item>
                <Checkbox
                  onChange={(e) => handleQuestionSelect(index, e.target.checked)}
                  checked={selectedQuestions.includes(index)}
                  style={{ marginRight: 8 }}
                />
                <div style={{ flex: 1 }}>
                  <strong>题目 {index + 1}:</strong> {question.content}
                  <div style={{ marginTop: 8 }}>
                    <small>
                      类型: {question.type} | 难度: {question.difficulty}
                      {question.language && ` | 语言: ${question.language}`}
                    </small>
                  </div>
                </div>
              </List.Item>
            )}
          />
        </Card>
      )}
    </div>
  )
}

export default AIGeneratePanel