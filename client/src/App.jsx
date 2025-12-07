import React, { useState, useEffect } from 'react'
import { Layout, Menu, Button, message } from 'antd'
import {
  MenuUnfoldOutlined,
  MenuFoldOutlined,
  HomeOutlined,
  DatabaseOutlined,
  PlusOutlined
} from '@ant-design/icons'
import { BrowserRouter as Router, Routes, Route, Link } from 'react-router-dom'
import HomePage from './pages/HomePage'
import QuestionManagement from './pages/QuestionManagement'
import CreateQuestionModal from './components/CreateQuestionModal'
import './App.css'

const { Header, Sider, Content } = Layout

function App() {
  const [collapsed, setCollapsed] = useState(false)
  const [isModalVisible, setIsModalVisible] = useState(false)
  const [modalMode, setModalMode] = useState('create') // 'create' or 'edit'

  const toggleCollapsed = () => {
    setCollapsed(!collapsed)
  }

  const handleCreateQuestion = () => {
    setModalMode('create')
    setIsModalVisible(true)
  }

  return (
    <Router>
      <Layout style={{ minHeight: '100vh' }}>
        <Sider trigger={null} collapsible collapsed={collapsed}>
          <div className="logo">题库系统</div>
          <Menu theme="dark" mode="inline" defaultSelectedKeys={['1']}>
            <Menu.Item key="1" icon={<HomeOutlined />}>
              <Link to="/">学习心得</Link>
            </Menu.Item>
            <Menu.Item key="2" icon={<DatabaseOutlined />}>
              <Link to="/questions">题库管理</Link>
            </Menu.Item>
          </Menu>
        </Sider>
        <Layout>
          <Header style={{ background: '#fff', padding: 0 }}>
            <Button
              type="text"
              icon={collapsed ? <MenuUnfoldOutlined /> : <MenuFoldOutlined />}
              onClick={toggleCollapsed}
              style={{ fontSize: '16px', width: 64, height: 64 }}
            />
            <Button
              type="primary"
              icon={<PlusOutlined />}
              onClick={handleCreateQuestion}
              style={{ float: 'right', margin: '16px 24px 0 0' }}
            >
              出题
            </Button>
          </Header>
          <Content style={{ margin: '24px 16px', padding: 24, background: '#fff' }}>
            <Routes>
              <Route path="/" element={<HomePage />} />
              <Route path="/questions" element={<QuestionManagement />} />
            </Routes>
          </Content>
        </Layout>
      </Layout>

      <CreateQuestionModal
        visible={isModalVisible}
        onClose={() => setIsModalVisible(false)}
        mode={modalMode}
      />
    </Router>
  )
}

export default App