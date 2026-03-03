import { useState, useEffect } from 'react'
import { getMenusByMonth, type MenuItem } from '../api'

export default function MenuList() {
  const [month, setMonth] = useState('2026-04')
  const [menus, setMenus] = useState<MenuItem[]>([])
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')

  useEffect(() => {
    setLoading(true)
    setError('')
    getMenusByMonth(month)
      .then(setMenus)
      .catch((e) => setError(e.message))
      .finally(() => setLoading(false))
  }, [month])

  return (
    <div className="card">
      <h1>献立一覧</h1>
      <div className="form-group">
        <label>対象月</label>
        <input
          type="month"
          value={month}
          onChange={(e) => setMonth(e.target.value)}
        />
      </div>
      {error && <p className="error">{error}</p>}
      {loading && <p>読み込み中...</p>}
      {!loading && menus.length > 0 && (
        <table>
          <thead>
            <tr>
              <th>日付</th>
              <th>曜日</th>
              <th>主食</th>
              <th>主菜</th>
              <th>副菜</th>
              <th>汁物</th>
              <th>デザート</th>
              <th>エネルギー</th>
            </tr>
          </thead>
          <tbody>
            {menus.map((m) => (
              <tr key={m.date}>
                <td>{m.date}</td>
                <td>{m.weekday}</td>
                <td>{m.staple}</td>
                <td>{m.main}</td>
                <td>{m.side}</td>
                <td>{m.soup}</td>
                <td>{m.dessert}</td>
                <td>{m.energy} kcal</td>
              </tr>
            ))}
          </tbody>
        </table>
      )}
      {!loading && menus.length === 0 && !error && <p>献立が登録されていません。</p>}
    </div>
  )
}
