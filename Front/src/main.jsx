import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import App from './Pages/Home'

createRoot(document.getElementById('root')).render(
  <StrictMode>
    <App />
  </StrictMode>,
)
