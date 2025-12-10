import React, { useState } from 'react'
import { BrowserRouter as Router, Routes, Route, Link, useLocation } from 'react-router-dom'
import { HomeOutlined, DatabaseOutlined, MenuFoldOutlined, MenuUnfoldOutlined, LeftCircleOutlined, LeftSquareOutlined } from '@ant-design/icons'
import HomePage from './pages/Homepage'
import QuestionManagement from './pages/QuestionManagement'
import './App.css'

const NavItem = ({ to, icon, label, active, collapsed }) => (
  <Link to={to} className={`nav-item ${active ? 'active' : ''} ${collapsed ? 'collapsed' : ''}`}>
    {icon}
    {!collapsed && <span>{label}</span>}
  </Link>
)

const AppShell = () => {
  const location = useLocation()
  const activePath = location.pathname
  const [collapsed, setCollapsed] = useState(false)

  return (
    <div className="app-layout">
      <div className="top-bar">
        <div className="brand">
          <div className="brand-icon">K</div>
          <div className="brand-text">
            金山办公
            <small>KINGSOFT OFFICE</small>
          </div>
        </div>
        <div >华中农业大学 付波 大作业</div>
      </div>

      <div className="layout-main">
        <aside className={`sidebar ${collapsed ? 'collapsed' : ''}`}>
          <div className="sidebar-nav">
            <NavItem
              to="/"
              icon={<HomeOutlined />}
              label="学习心得"
              active={activePath === '/'}
              collapsed={collapsed}
            />
            <NavItem
              to="/questions"
              icon={<DatabaseOutlined />}
              label="题库管理"
              active={activePath.startsWith('/questions')}
              collapsed={collapsed}
            />
          </div>
          <button
            className="collapse-toggle"
            type="button"
            onClick={() => setCollapsed(!collapsed)}
          >
            {collapsed ? <LeftSquareOutlined /> : <LeftSquareOutlined />}
          </button>
        </aside>

        <main className="content">
          <Routes>
            <Route path="/" element={<HomePage />} />
            <Route path="/questions" element={<QuestionManagement />} />
          </Routes>
        </main>
      </div>
    </div>
  )
}

const App = () => (
  <Router>
    <AppShell />
  </Router>
)

export default App