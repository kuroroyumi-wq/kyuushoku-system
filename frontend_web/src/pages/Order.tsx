import { useState } from 'react'
import { getOrder, downloadExport } from '../api'

export default function Order() {
  const [start, setStart] = useState('2026-04-01')
  const [end, setEnd] = useState('2026-04-30')
  const [people, setPeople] = useState(120)
  const [data, setData] = useState<Awaited<ReturnType<typeof getOrder>> | null>(null)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setLoading(true)
    setError('')
    setData(null)
    getOrder(start, end, people)
      .then(setData)
      .catch((e) => setError(e.message))
      .finally(() => setLoading(false))
  }

  const handleExport = async () => {
    const month = start.slice(0, 7)
    try {
      const blob = await downloadExport(month)
      const url = URL.createObjectURL(blob)
      const a = document.createElement('a')
      a.href = url
      a.download = `献立表_${month.replace('-', '年')}月.xlsx`
      a.click()
      URL.revokeObjectURL(url)
    } catch (e) {
      setError(String(e))
    }
  }

  return (
    <div className="card">
      <h1>発注集計</h1>
      <form onSubmit={handleSubmit}>
        <div className="form-group">
          <label>開始日</label>
          <input type="date" value={start} onChange={(e) => setStart(e.target.value)} required />
        </div>
        <div className="form-group">
          <label>終了日</label>
          <input type="date" value={end} onChange={(e) => setEnd(e.target.value)} required />
        </div>
        <div className="form-group">
          <label>人数</label>
          <input
            type="number"
            min={1}
            value={people}
            onChange={(e) => setPeople(Number(e.target.value))}
            required
          />
        </div>
        <button type="submit" disabled={loading}>
          {loading ? '集計中...' : '集計'}
        </button>
      </form>
      {error && <p className="error">{error}</p>}
      {data && (
        <div className="card" style={{ marginTop: 16 }}>
          <h2>発注集計（{data.start} 〜 {data.end}、{data.people}人分）</h2>
          <table>
            <thead>
              <tr>
                <th>食材名</th>
                <th>総使用量(g)</th>
              </tr>
            </thead>
            <tbody>
              {data.items.map((item) => (
                <tr key={item.ingredient_id}>
                  <td>{item.name}</td>
                  <td>{item.total_g}</td>
                </tr>
              ))}
            </tbody>
          </table>
          <button onClick={handleExport} style={{ marginTop: 16 }}>
            献立表Excelをダウンロード
          </button>
        </div>
      )}
    </div>
  )
}
