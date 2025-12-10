import React from 'react'
import { Card, Typography, Button, Empty } from 'antd'
import { PlusOutlined } from '@ant-design/icons'

const { Title, Paragraph } = Typography

const QuestionBank: React.FC = () => {
  return (
    <Card
      title="题库管理"
      style={{
        height: '100%',
        border: 'none',
        boxShadow: 'none',
      }}
      bodyStyle={{
        padding: 0,
        height: 'calc(100% - 57px)',
        overflow: 'auto',
      }}
      extra={
        <Button type="primary" icon={<PlusOutlined />}>
          新增题目
        </Button>
      }
    >
      <div style={{ padding: '24px' }}>
        <Title level={4}>题库管理页面</Title>
        <Paragraph>
          题库管理功能正在开发中，将包含以下功能：
        </Paragraph>
        
        <div style={{ marginTop: '24px' }}>
          <Empty
            description="功能开发中"
            image={Empty.PRESENTED_IMAGE_SIMPLE}
          />
        </div>
        
        <div style={{ marginTop: '32px' }}>
          <Title level={5}>规划功能：</Title>
          <ul style={{ lineHeight: '2', paddingLeft: '20px' }}>
            <li>题目分类管理</li>
            <li>题目增删改查</li>
            <li>题目导入导出</li>
            <li>题目批量操作</li>
            <li>题目搜索和筛选</li>
            <li>题目难度分级</li>
          </ul>
        </div>
      </div>
    </Card>
  )
}

export default QuestionBank