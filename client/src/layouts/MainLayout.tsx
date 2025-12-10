import React, { useState, useEffect } from 'react'
import { Layout, Menu, Button, theme, Avatar } from 'antd'
import {
  MenuFoldOutlined,
  MenuUnfoldOutlined,
  ReadOutlined,
  DatabaseOutlined,
  UserOutlined,
} from '@ant-design/icons'
import { Outlet, useNavigate, useLocation } from 'react-router-dom'

const { Header, Sider, Content } = Layout

const MainLayout: React.FC = () => {
  const [collapsed, setCollapsed] = useState(false)
  const [mounted, setMounted] = useState(false)
  const navigate = useNavigate()
  const location = useLocation()
  const {
    token: { colorBgContainer, borderRadiusLG },
  } = theme.useToken()

  // 防止 SSR 样式闪烁
  useEffect(() => {
    setMounted(true)
  }, [])

  // 菜单项
  const menuItems = [
    {
      key: '/study',
      icon: <ReadOutlined />,
      label: '学习心得',
    },
    {
      key: '/question-bank',
      icon: <DatabaseOutlined />,
      label: '题库管理',
    },
  ]

  // 处理菜单点击
  const handleMenuClick = (e: any) => {
    navigate(e.key)
  }

  // 获取当前选中的菜单项
  const getSelectedKey = () => {
    if (location.pathname === '/' || location.pathname === '/study') {
      return ['/study']
    }
    return [location.pathname]
  }

  // 折叠按钮样式
  const triggerStyle: React.CSSProperties = {
    fontSize: '16px',
    width: 40,
    height: 40,
    display: 'flex',
    alignItems: 'center',
    justifyContent: 'center',
    borderRadius: '50%',
    backgroundColor: 'rgba(255, 255, 255, 0.1)',
    border: 'none',
    cursor: 'pointer',
  }

  if (!mounted) {
    return null
  }

  return (
    <Layout style={{ minHeight: '100vh', backgroundColor: '#000' }}>
      {/* 左侧黑色导航栏 */}
      <Sider
        trigger={null}
        collapsible
        collapsed={collapsed}
        width={240}
        collapsedWidth={80}
        style={{
          backgroundColor: '#000',
          overflow: 'auto',
          height: '100vh',
          position: 'fixed',
          left: 0,
          top: 0,
          bottom: 0,
          zIndex: 1000,
          borderRight: 'none',
        }}
      >
        {/* 左上角logo和文字区域 */}
        <div
          style={{
            height: '64px',
            display: 'flex',
            alignItems: 'center',
            padding: collapsed ? '0 20px' : '0 24px',
            borderBottom: '1px solid rgba(255, 255, 255, 0.1)',
            justifyContent: collapsed ? 'center' : 'flex-start',
          }}
        >
          {collapsed ? (
            <Avatar
              size="large"
              icon={<UserOutlined />}
              style={{ backgroundColor: '#1890ff' }}
            />
          ) : (
            <>
              <Avatar
                size="large"
                icon={<UserOutlined />}
                style={{ backgroundColor: '#1890ff', marginRight: '12px' }}
              />
              <div style={{ color: '#fff', lineHeight: '1.2' }}>
                <div style={{ fontSize: '16px', fontWeight: 'bold' }}>
                  武汉科技大学
                </div>
                <div style={{ fontSize: '12px', opacity: 0.8 }}>张三</div>
              </div>
            </>
          )}
        </div>

        {/* 导航菜单 */}
        <Menu
          theme="dark"
          mode="inline"
          selectedKeys={getSelectedKey()}
          items={menuItems}
          onClick={handleMenuClick}
          style={{
            backgroundColor: '#000',
            borderRight: 'none',
            marginTop: '16px',
          }}
        />

        {/* 左下角折叠按钮 */}
        <div
          style={{
            position: 'absolute',
            bottom: '16px',
            left: 0,
            right: 0,
            display: 'flex',
            justifyContent: 'center',
          }}
        >
          <Button
            type="text"
            icon={
              collapsed ? (
                <MenuUnfoldOutlined style={{ color: '#fff' }} />
              ) : (
                <MenuFoldOutlined style={{ color: '#fff' }} />
              )
            }
            onClick={() => setCollapsed(!collapsed)}
            style={triggerStyle}
          />
        </div>
      </Sider>

      {/* 主内容区域 */}
      <Layout
        style={{
          marginLeft: collapsed ? 80 : 240,
          transition: 'all 0.2s',
          minHeight: '100vh',
          backgroundColor: '#000',
        }}
      >
        {/* 顶部黑色区域 */}
        <Header
          style={{
            padding: 0,
            backgroundColor: '#000',
            height: '64px',
            position: 'sticky',
            top: 0,
            zIndex: 999,
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'space-between',
            paddingLeft: '24px',
            paddingRight: '24px',
            borderBottom: 'none',
            boxShadow: '0 2px 8px rgba(0, 0, 0, 0.15)',
          }}
        >
          <div
            style={{
              color: '#fff',
              fontSize: '18px',
              fontWeight: 'bold',
              display: 'flex',
              alignItems: 'center',
            }}
          >
            金山办公
            <span style={{ margin: '0 8px', opacity: 0.6 }}>|</span>
            武汉科技大学
            <span style={{ margin: '0 8px', opacity: 0.6 }}>|</span>
            张三
            <span style={{ margin: '0 8px', opacity: 0.6 }}>|</span>
            大作业
          </div>
        </Header>

        {/* 内容区域 */}
        <Content
          style={{
            margin: '24px',
            padding: 0,
            minHeight: 'calc(100vh - 112px)',
            background: colorBgContainer,
            borderRadius: borderRadiusLG,
            backgroundColor: '#fff',
            overflow: 'auto',
            boxShadow: '0 2px 8px rgba(0, 0, 0, 0.1)',
          }}
        >
          <Outlet />
        </Content>
      </Layout>
    </Layout>
  )
}

export default MainLayout