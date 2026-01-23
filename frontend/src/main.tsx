import ReactDOM from 'react-dom/client'
import { RouterProvider } from 'react-router-dom'
import { ConfigProvider, theme as antdTheme } from 'antd'
import zhCN from 'antd/locale/zh_CN'
import router from './router'
import { useThemeStore } from './stores/themeStore'
import './index.css'

function App() {
  const actualTheme = useThemeStore((state) => state.actualTheme)

  return (
    <ConfigProvider
      locale={zhCN}
      theme={{
        algorithm: actualTheme === 'dark' ? antdTheme.darkAlgorithm : antdTheme.defaultAlgorithm,
        token: {
          colorPrimary: '#3a84ff',
          borderRadius: 6,
        },
        components: {
          Layout: {
            bodyBg: actualTheme === 'dark' ? '#0a0e1a' : '#ffffff',
            headerBg: actualTheme === 'dark' ? '#141824' : '#ffffff',
            siderBg: actualTheme === 'dark' ? '#141824' : '#ffffff',
          },
        },
      }}
    >
      <RouterProvider router={router} />
    </ConfigProvider>
  )
}

ReactDOM.createRoot(document.getElementById('root')!).render(<App />)

