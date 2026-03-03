import { useState } from 'react'
import { getCalc } from '../api'

export default function Nutrition() {
  const [date, setDate] = useState('2026-04-01')
  const [data, setData] = useState<Awaited<ReturnType<typeof getCalc>> | null>(null)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setLoading(true)
    setError('')
    setData(null)
    getCalc(date)
      .then(setData)
      .catch((e) => setError(e.message))
      .finally(() => setLoading(false))
  }

  return (
    <div className="card">
      <h1>栄養計算</h1>
      <form onSubmit={handleSubmit}>
        <div className="form-group">
          <label>日付</label>
          <input
            type="date"
            value={date}
            onChange={(e) => setDate(e.target.value)}
            required
          />
        </div>
        <button type="submit" disabled={loading}>
          {loading ? '計算中...' : '計算'}
        </button>
      </form>
      {error && <p className="error">{error}</p>}
      {data && (
        <div className="card" style={{ marginTop: 16 }}>
          <h2>献立の栄養価（{data.date}）</h2>
          <table>
            <tbody>
              <tr>
                <td>エネルギー</td>
                <td>{data.energy} kcal</td>
                <td>（目標 {data.targets.energy} kcal）</td>
              </tr>
              <tr>
                <td>たんぱく質</td>
                <td>{data.protein} g</td>
                <td>（目標 {data.targets.protein} g）</td>
              </tr>
              <tr>
                <td>脂質</td>
                <td>{data.fat} g</td>
              </tr>
              <tr>
                <td>炭水化物</td>
                <td>{data.carbohydrate} g</td>
              </tr>
              <tr>
                <td>食塩相当量</td>
                <td>{data.salt} g</td>
              </tr>
            </tbody>
          </table>
        </div>
      )}
    </div>
  )
}
