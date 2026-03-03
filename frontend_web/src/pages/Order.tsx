import { useState } from 'react'
import { getOrder, downloadExport } from '../api'

export default function Order() {
  const [mode, setMode] = useState<'daily' | 'period'>('daily')
  const [date, setDate] = useState('2026-04-01')
  const [start, setStart] = useState('2026-04-01')
  const [end, setEnd] = useState('2026-04-30')
  const [people, setPeople] = useState(120)
  const [excludeCondiments, setExcludeCondiments] = useState(false)
  const [data, setData] = useState<Awaited<ReturnType<typeof getOrder>> | null>(null)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setLoading(true)
    setError('')
    setData(null)
    const s = mode === 'daily' ? date : start
    const e_ = mode === 'daily' ? date : end
    getOrder(s, e_, people, excludeCondiments)
      .then(setData)
      .catch((e) => setError(e.message))
      .finally(() => setLoading(false))
  }

  const handleExport = async () => {
    const month = (mode === 'daily' ? date : start).slice(0, 7)
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

  const title = data
    ? (mode === 'daily' ? `${data.start}（${data.people}人分）` : `${data.start} 〜 ${data.end}（${data.people}人分）`)
    : ''

  return (
    <div className="card">
      <h1>発注集計</h1>
      <form onSubmit={handleSubmit}>
        <div className="form-group">
          <label>集計モード</label>
          <div style={{ display: 'flex', gap: 16, marginBottom: 8 }}>
            <label style={{ display: 'flex', alignItems: 'center', gap: 4 }}>
              <input
                type="radio"
                name="mode"
                checked={mode === 'daily'}
                onChange={() => setMode('daily')}
              />
              1日単位（日別発注）
            </label>
            <label style={{ display: 'flex', alignItems: 'center', gap: 4 }}>
              <input
                type="radio"
                name="mode"
                checked={mode === 'period'}
                onChange={() => setMode('period')}
              />
              期間で集計
            </label>
          </div>
        </div>

        {mode === 'daily' ? (
          <div className="form-group">
            <label>日付</label>
            <input type="date" value={date} onChange={(e) => setDate(e.target.value)} required />
          </div>
        ) : (
          <>
            <div className="form-group">
              <label>開始日</label>
              <input type="date" value={start} onChange={(e) => setStart(e.target.value)} required />
            </div>
            <div className="form-group">
              <label>終了日</label>
              <input type="date" value={end} onChange={(e) => setEnd(e.target.value)} required />
            </div>
          </>
        )}

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

        <div className="form-group">
          <label style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
            <input
              type="checkbox"
              checked={excludeCondiments}
              onChange={(e) => setExcludeCondiments(e.target.checked)}
            />
            調味料を除外（別途発注する場合）
          </label>
        </div>

        <button type="submit" disabled={loading}>
          {loading ? '集計中...' : '集計'}
        </button>
      </form>

      {error && <p className="error">{error}</p>}

      {data && (
        <div className="card" style={{ marginTop: 16 }}>
          <h2>発注集計（{title}）</h2>
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
