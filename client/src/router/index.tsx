import { createBrowserRouter } from 'react-router-dom'
import MainLayout from '@/layouts/MainLayout'
import StudyNotes from '@/pages/StudyNotes'
import QuestionBank from '@/pages/QuestionBank'

const router = createBrowserRouter([
  {
    path: '/',
    element: <MainLayout />,
    children: [
      {
        index: true,
        element: <StudyNotes />,
      },
      {
        path: 'study',
        element: <StudyNotes />,
      },
      {
        path: 'question-bank',
        element: <QuestionBank />,
      },
    ],
  },
])

export default router