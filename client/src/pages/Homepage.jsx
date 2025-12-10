import React, { useState, useEffect } from 'react'
import { Card, Spin } from 'antd'
import ReactMarkdown from 'react-markdown'
import axios from 'axios'
import './HomePage.css'

const HomePage = () => {
  const [content, setContent] = useState('')
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    fetchLearningNote()
  }, [])

  const fetchLearningNote = async () => {
    try {
      const response = await axios.get('/api/learning-note')
      setContent(response.data.content)
    } catch (error) {
      console.error('Failed to fetch learning note:', error)
      setContent('# 学习心得\n\n加载失败，请检查服务器连接。')
    } finally {
      setLoading(false)
    }
  }

  if (loading) {
    return (
      <div className="loading-container">
        <Spin size="large" />
      </div>
    )
  }

  return (
    <div className="card">
      <div className="page-title">学习心得</div>
      <div className="home-placeholder">
        {content ? (
          <div className="markdown-container">
            <ReactMarkdown>{content}</ReactMarkdown>
          </div>
        ) : (
          '这里放学习心得，至少100字'
        )}
      </div>
    </div>
  )
}

export default HomePage