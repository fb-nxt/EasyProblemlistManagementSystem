import React, { useState, useEffect } from 'react'
import {
  Modal,
  Form,
  Input,
  Select,
  Button,
  Space,
  Radio,
  message,
  Tabs,
  Checkbox,
  InputNumber,
  Row,
  Col,
  Card,
  List
} from 'antd'
import axios from 'axios'

const { Option } = Select
const { TextArea } = Input
const { TabPane } = Tabs

const CreateQuestionModal = ({ visible, onClose, mode, question, defaultTab = 'manual' }) => {
  const [form] = Form.useForm()
  const [aiForm] = Form.useForm()
  const [activeTab, setActiveTab] = useState(defaultTab)
  const [loading, setLoading] = useState(false)
  const [aiLoading, setAiLoading] = useState(false)
  const [aiGeneratedQuestions, setAiGeneratedQuestions] = useState([])
  const [selectedAiQuestions, setSelectedAiQuestions] = useState([])

  useEffect(() => {
    if (mode === 'edit' && question) {
      const options = question.options || {}
      const isMultiple = question.type === 'multiple_choice'
      form.setFieldsValue({
        type: question.type,
        content: question.content,
        difficulty: question.difficulty,
        language: question.language,
        answerSingle: !isMultiple ? question.answer : undefined,
        answerMultiple: isMultiple && question.answer ? question.answer.split(',') : [],
        optionsA: options.A,
        optionsB: options.B,
        optionsC: options.C,
        optionsD: options.D
      })
    } else {
      form.resetFields()
    }
  }, [mode, question, form])

  // 当visible或defaultTab变化时，更新activeTab
  useEffect(() => {
    if (visible) {
      setActiveTab(defaultTab)
    }
  }, [visible, defaultTab])

  // 切换tab时重置AI相关状态
  useEffect(() => {
    if (activeTab === 'ai') {
      setAiGeneratedQuestions([])
      setSelectedAiQuestions([])
      aiForm.resetFields()
    }
  }, [activeTab, aiForm])

  const handleSubmit = async (values) => {
    setLoading(true)
    try {
      const payload = {
        type: values.type,
        content: values.content,
        difficulty: values.difficulty,
        language: values.language || ''
      }

      if (values.type !== 'programming') {
        payload.options = {
          A: values.optionsA || '',
          B: values.optionsB || '',
          C: values.optionsC || '',
          D: values.optionsD || ''
        }
        payload.answer =
          values.type === 'multiple_choice'
            ? (values.answerMultiple || []).join(',')
            : values.answerSingle || ''
      }

      console.log('提交的payload:', payload)

      if (mode === 'edit') {
        const response = await axios.put(`/api/questions/${question.id}`, payload)
        console.log('更新响应:', response.data)
        message.success('更新成功')
      } else {
        const response = await axios.post('/api/questions', payload)
        console.log('创建响应:', response.data)
        message.success('创建成功')
      }
      onClose()
    } catch (error) {
      console.error('提交错误:', error)
      const errorMessage = error.response?.data?.error || error.message || '操作失败'
      message.error(`${mode === 'edit' ? '更新失败' : '创建失败'}: ${errorMessage}`)
    } finally {
      setLoading(false)
    }
  }

  const handleAiGenerate = async (values) => {
    setAiLoading(true)
    try {
      const response = await axios.post('/api/ai/generate', values)
      setAiGeneratedQuestions(response.data.data)
      setSelectedAiQuestions([])
      message.success('题目生成成功')
    } catch (error) {
      message.error('题目生成失败')
      console.error('Error generating questions:', error)
    } finally {
      setAiLoading(false)
    }
  }

  const handleAiQuestionSelect = (index, checked) => {
    const next = checked
      ? [...selectedAiQuestions, index]
      : selectedAiQuestions.filter((idx) => idx !== index)
    setSelectedAiQuestions(next)
  }

  const handleAddAiQuestions = async () => {
    if (selectedAiQuestions.length === 0) {
      message.warning('请先选择要添加的题目')
      return
    }

    setLoading(true)
    try {
      console.log('准备添加的题目:', selectedAiQuestions)
      // 批量添加AI生成的题目
      const selectedQuestionObjects = selectedAiQuestions.map(idx => aiGeneratedQuestions[idx])
      const promises = selectedQuestionObjects.map(q => {
        console.log('添加题目:', q)
        return axios.post('/api/questions', q)
      })
      const results = await Promise.all(promises)
      console.log('添加结果:', results)
      message.success(`成功添加 ${selectedAiQuestions.length} 道题目`)
      onClose()
    } catch (error) {
      console.error('添加题目错误:', error)
      const errorMessage = error.response?.data?.error || error.message || '添加失败'
      message.error(`添加题目失败: ${errorMessage}`)
    } finally {
      setLoading(false)
    }
  }

  return (
    <Modal
      title={mode === 'edit' ? '编辑题目' : '出题'}
      open={visible}
      onCancel={onClose}
      width={activeTab === 'ai' ? 1200 : 800}
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
              difficulty: 'medium',
              language: 'Go',
              answerMultiple: []
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
              rules={[{ required: true, message: '请选择编程语言' }]}
            >
              <Select placeholder="请选择编程语言">
                <Option value="Go">Go</Option>
                <Option value="JavaScript">JavaScript</Option>
              </Select>
            </Form.Item>

            <Form.Item
              shouldUpdate={(prevValues, curValues) => prevValues.type !== curValues.type}
            >
              {({ getFieldValue }) => {
                const type = getFieldValue('type')
                if (type === 'programming') return null

                return (
                  <>
                    <Form.Item
                      name="optionsA"
                      label="A"
                      rules={[{ required: true, message: '请输入选项A' }]}
                    >
                      <Input placeholder="请输入选项内容" />
                    </Form.Item>
                    <Form.Item
                      name="optionsB"
                      label="B"
                      rules={[{ required: true, message: '请输入选项B' }]}
                    >
                      <Input placeholder="请输入选项内容" />
                    </Form.Item>
                    <Form.Item
                      name="optionsC"
                      label="C"
                      rules={[{ required: true, message: '请输入选项C' }]}
                    >
                      <Input placeholder="请输入选项内容" />
                    </Form.Item>
                    <Form.Item
                      name="optionsD"
                      label="D"
                      rules={[{ required: true, message: '请输入选项D' }]}
                    >
                      <Input placeholder="请输入选项内容" />
                    </Form.Item>

                    {type === 'single_choice' ? (
                      <Form.Item
                        name="answerSingle"
                        label="答案"
                        rules={[{ required: true, message: '请选择正确答案' }]}
                      >
                        <Radio.Group>
                          <Radio value="A">A</Radio>
                          <Radio value="B">B</Radio>
                          <Radio value="C">C</Radio>
                          <Radio value="D">D</Radio>
                        </Radio.Group>
                      </Form.Item>
                    ) : (
                      <Form.Item
                        name="answerMultiple"
                        label="答案"
                        rules={[{ required: true, message: '请选择正确答案' }]}
                      >
                        <Checkbox.Group
                          options={[
                            { label: 'A', value: 'A' },
                            { label: 'B', value: 'B' },
                            { label: 'C', value: 'C' },
                            { label: 'D', value: 'D' }
                          ]}
                        />
                      </Form.Item>
                    )}
                  </>
                )
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
          <Row gutter={24}>
            {/* 左侧：参数配置 */}
            <Col span={10}>
              <Form
                form={aiForm}
                layout="vertical"
                onFinish={handleAiGenerate}
                initialValues={{
                  type: 'single_choice',
                  count: 3,
                  difficulty: 'medium',
                  language: 'Go'
                }}
              >
                <Form.Item
                  name="type"
                  label="题型"
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
                  label={
                    <span>
                      题目数量
                      <span style={{ marginLeft: 8, fontSize: '12px', color: '#999', fontWeight: 'normal' }}>
                        (最小1,最大10)
                      </span>
                    </span>
                  }
                  rules={[{ required: true, message: '请输入题目数量' }]}
                >
                  <InputNumber min={1} max={10} style={{ width: '100%' }} />
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
                  rules={[{ required: true, message: '请选择编程语言' }]}
                >
                  <Select placeholder="请选择编程语言">
                    <Option value="Go">Go</Option>
                    <Option value="JavaScript">JavaScript</Option>
                  </Select>
                </Form.Item>

                <Form.Item name="topic" label="题目主题">
                  <TextArea rows={2} placeholder="输入题目主题或关键词（可选）" />
                </Form.Item>

                <Form.Item>
                  <Button type="primary" htmlType="submit" loading={aiLoading} block size="large">
                    生成并预览题库
                  </Button>
                </Form.Item>
              </Form>
            </Col>

            {/* 右侧：题目预览 */}
            <Col span={14}>
              <Card
                title="生成的题目预览"
                extra={
                  aiGeneratedQuestions.length > 0 && (
                    <Button
                      type="primary"
                      onClick={handleAddAiQuestions}
                      loading={loading}
                      disabled={selectedAiQuestions.length === 0}
                    >
                      添加选中题目到题库 ({selectedAiQuestions.length})
                    </Button>
                  )
                }
                style={{
                  height: '100%',
                  display: 'flex',
                  flexDirection: 'column'
                }}
                bodyStyle={{
                  flex: 1,
                  overflow: 'auto',
                  padding: aiGeneratedQuestions.length === 0 ? '0' : '24px'
                }}
              >
                {aiGeneratedQuestions.length === 0 ? (
                  <div style={{
                    textAlign: 'center',
                    padding: '60px 20px',
                    color: '#999',
                    fontSize: '14px',
                    border: '1px dashed #d9d9d9',
                    borderRadius: '4px',
                    height: '100%',
                    display: 'flex',
                    alignItems: 'center',
                    justifyContent: 'center'
                  }}>
                    AI生成区域
                  </div>
                ) : (
                  <div style={{ maxHeight: '500px', overflowY: 'auto' }}>
                    <List
                      dataSource={aiGeneratedQuestions}
                      renderItem={(question, index) => (
                        <List.Item
                          style={{
                            padding: '16px',
                            border: selectedAiQuestions.includes(index)
                              ? '2px solid #1890ff'
                              : '1px solid #e8e8e8',
                            borderRadius: '6px',
                            marginBottom: '12px',
                            backgroundColor: selectedAiQuestions.includes(index)
                              ? '#e6f7ff'
                              : '#fafafa',
                            cursor: 'pointer',
                            transition: 'all 0.3s'
                          }}
                          onClick={() => handleAiQuestionSelect(index, !selectedAiQuestions.includes(index))}
                        >
                          <Checkbox
                            checked={selectedAiQuestions.includes(index)}
                            onChange={(e) => {
                              e.stopPropagation()
                              handleAiQuestionSelect(index, e.target.checked)
                            }}
                            style={{ marginRight: 16, marginTop: 4 }}
                          />
                          <div style={{ flex: 1 }}>
                            <div style={{ marginBottom: 12, fontSize: '15px', lineHeight: '1.6' }}>
                              <strong style={{ color: '#1890ff' }}>题目 {index + 1}:</strong> {question.content}
                            </div>
                            {question.options && (
                              <div style={{ marginLeft: 0, marginBottom: 10, fontSize: '14px', color: '#666' }}>
                                {Object.entries(question.options).map(([key, value]) => (
                                  <div key={key} style={{ marginBottom: 6, paddingLeft: '20px' }}>
                                    <strong style={{ color: '#333' }}>{key}.</strong> {value}
                                  </div>
                                ))}
                              </div>
                            )}
                            {question.answer && (
                              <div style={{ marginLeft: 0, marginBottom: 8, fontSize: '14px', color: '#52c41a', fontWeight: '500' }}>
                                <strong>答案:</strong> {question.answer}
                              </div>
                            )}
                            <div style={{ marginTop: 10, fontSize: '12px', color: '#999', paddingTop: '8px', borderTop: '1px solid #f0f0f0' }}>
                              <span>类型: {question.type === 'single_choice' ? '单选题' :
                                question.type === 'multiple_choice' ? '多选题' : '编程题'}</span>
                              <span style={{ margin: '0 12px' }}>|</span>
                              <span>难度: {question.difficulty === 'easy' ? '简单' :
                                question.difficulty === 'medium' ? '中等' : '困难'}</span>
                              {question.language && (
                                <>
                                  <span style={{ margin: '0 12px' }}>|</span>
                                  <span>语言: {question.language}</span>
                                </>
                              )}
                            </div>
                          </div>
                        </List.Item>
                      )}
                    />
                  </div>
                )}
              </Card>
            </Col>
          </Row>
        </TabPane>
      </Tabs>
    </Modal>
  )
}

export default CreateQuestionModal