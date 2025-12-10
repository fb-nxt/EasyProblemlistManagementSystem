import React, { useState, useEffect } from 'react'
import ReactMarkdown from 'react-markdown'
import { Card, Spin, Alert } from 'antd'
import remarkGfm from 'remark-gfm'
import 'github-markdown-css/github-markdown.css'

const StudyNotes: React.FC = () => {
  const [markdownContent, setMarkdownContent] = useState<string>('')
  const [loading, setLoading] = useState<boolean>(true)
  const [error, setError] = useState<string>('')

  useEffect(() => {
    const fetchMarkdown = async () => {
      try {
        setLoading(true)
        setError('')
        
        // 尝试从不同路径获取 markdown 文件
        const possiblePaths = [
          '../学习心得.md',
          '/学习心得.md',
          './学习心得.md',
        ]
        
        let success = false
        
        for (const path of possiblePaths) {
          try {
            const response = await fetch(path)
            if (response.ok) {
              const text = await response.text()
              setMarkdownContent(text)
              success = true
              break
            }
          } catch (e) {
            // 继续尝试下一个路径
            continue
          }
        }
        
        if (!success) {
          throw new Error('无法加载学习心得文件，请确保文件存在')
        }
      } catch (err: any) {
        console.error('加载学习心得文件失败:', err)
        setError(err.message || '加载失败')
        // 使用默认内容
        setMarkdownContent(`
# React + TypeScript 前端开发学习心得

## 前言
通过本次大作业的学习和实践，我对React前端开发有了更深入的理解。本次大作业以"金山办公"为主题，结合武汉科技大学的课程要求，让我从零开始构建了一个完整的前端项目。

## 技术栈掌握情况

### 1. React组件化开发
学会了如何将UI拆分为独立的、可复用的组件，提高了代码的复用性和可维护性。通过创建布局组件、页面组件和功能组件，理解了组件之间的数据传递和通信机制。

### 2. TypeScript类型系统
通过TypeScript的静态类型检查，减少了运行时错误，提高了代码质量。学会了如何定义接口、类型别名，以及如何使用泛型来提高代码的灵活性。

### 3. React Hooks的使用
熟练掌握了useState、useEffect、useContext等常用Hook，能够编写函数式组件并管理状态和副作用。特别是useEffect的使用，让我理解了组件的生命周期和副作用管理。

### 4. React Router路由管理
使用React Router实现了页面路由和导航功能，理解了单页应用的路由原理。掌握了嵌套路由、动态路由和编程式导航等高级功能。

### 5. Ant Design组件库
学会了使用Ant Design快速构建美观、一致的用户界面。掌握了Layout、Menu、Card、Form等常用组件的使用，以及如何自定义主题样式。

### 6. Vite构建工具
体验了Vite的快速启动和热更新，了解了现代前端构建工具的优势。掌握了Vite的基本配置，包括别名配置、代理配置和构建优化。

## 项目实践收获

### 布局设计
通过实现黑色导航栏和白色内容区域的布局，加深了对CSS布局和定位的理解。特别是固定定位和边距计算的应用，让我对页面布局有了更深刻的认识。

### 状态管理
在项目中合理使用状态管理，包括组件的本地状态和通过Context共享的全局状态。理解了状态提升的概念和适用场景。

### 路由配置
实现了基于React Router的路由系统，支持嵌套路由和懒加载。通过路由守卫和权限控制，提高了应用的安全性。

### 响应式设计
考虑到不同设备的显示效果，实现了基本的响应式布局。通过媒体查询和Ant Design的响应式工具，确保了应用在移动设备上的可用性。

## 遇到的问题与解决方案

### 1. 样式冲突问题
在使用Ant Design时，遇到了自定义样式与组件默认样式冲突的问题。通过使用CSS Modules和styled-components，解决了样式隔离的问题。

### 2. 路由配置复杂
在配置嵌套路由时，初始配置较为复杂。通过查阅官方文档和社区资源，最终简化了路由配置，提高了代码的可读性。

### 3. 构建优化
在项目构建时，遇到了打包体积过大的问题。通过代码分割、按需加载和Tree Shaking，有效减小了打包体积。

## 总结与展望

本次大作业不仅提升了我的编码能力，更重要的是培养了我解决实际问题的能力。从需求分析、技术选型到代码实现和优化，每一个环节都让我受益匪浅。

通过这次实践，我深刻理解了前端工程化的重要性。一个好的项目不仅要有漂亮的外观，更要有良好的代码结构、合理的性能优化和可维护性。

未来我将继续深入学习前端技术，探索更多前沿的技术栈，如Next.js、微前端架构等。同时也会关注性能优化、用户体验等方向，努力成为一名优秀的前端工程师。

## 致谢

感谢武汉科技大学提供的学习机会，感谢老师的悉心指导，也感谢同学们的帮助和支持。这次大作业让我将理论知识与实践相结合，为未来的职业发展奠定了坚实的基础。
        `)
      } finally {
        setLoading(false)
      }
    }

    fetchMarkdown()
  }, [])

  return (
    <Card
      title="学习心得"
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
    >
      {loading ? (
        <div style={{ textAlign: 'center', padding: '50px' }}>
          <Spin size="large" tip="加载中..." />
        </div>
      ) : error ? (
        <Alert
          message="加载失败"
          description={error}
          type="error"
          showIcon
          style={{ margin: '20px' }}
        />
      ) : (
        <div className="markdown-body">
          <ReactMarkdown remarkPlugins={[remarkGfm]}>
            {markdownContent}
          </ReactMarkdown>
        </div>
      )}
    </Card>
  )
}

export default StudyNotes