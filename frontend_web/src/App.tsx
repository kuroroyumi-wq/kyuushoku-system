import { Routes, Route, Link } from 'react-router-dom'
import MenuInput from './pages/MenuInput'
import Nutrition from './pages/Nutrition'
import Order from './pages/Order'
import BulkOrder from './pages/BulkOrder'
import MenuList from './pages/MenuList'

function App() {
  return (
    <>
      <nav>
        <Link to="/">献立一覧</Link>
        <Link to="/menu-input">献立入力</Link>
        <Link to="/nutrition">栄養計算</Link>
        <Link to="/order">発注集計</Link>
        <Link to="/bulk-order">まとめ買い</Link>
      </nav>
      <div className="container">
        <Routes>
          <Route path="/" element={<MenuList />} />
          <Route path="/menu-input" element={<MenuInput />} />
          <Route path="/nutrition" element={<Nutrition />} />
          <Route path="/order" element={<Order />} />
          <Route path="/bulk-order" element={<BulkOrder />} />
        </Routes>
      </div>
    </>
  )
}

export default App
