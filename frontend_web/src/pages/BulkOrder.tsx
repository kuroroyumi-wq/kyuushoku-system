import { useState } from 'react'
import { getBulkOrder } from '../api'
import type { BulkOrderItem } from '../api'

const CATEGORY_LABELS: Record<string, string> = {
  '米・主食': '米・主食',
  '調味料': '調味料',
  '乾物・海藻': '乾物・海藻',
  '備蓄品': '備蓄品',
}

export default function BulkOrder() {
  const [start, setStart] = useState('2026-04-01')
  const [end, setEnd] = useState('2026-04-30')
  const [people, setPeople] = useState(120)
  const [data, setData] = useState<Awaited<ReturnType<typeof getBulkOrder>> | null>(null)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setLoading(true)
    setError('')
    setData(null)
    getBulkOrder(start, end, people)
      .then(setData)
      .catch((e) => setError(e.message))
      .finally(() => setLoading(false))
  }

  const byCategory = (items: BulkOrderItem[]) => {
    const map: Record<string, BulkOrderItem[]> = {}
    for (const item of items) {
      const cat = item.bulk_category || 'その他'
      if (!map[cat]) map[cat] = []
      map[cat].push(item)
    }
    return map
  }

  return (
    <div className="card">
      <h1>まとめ買い</h1>
      <p style={{ color: '#666', marginBottom: 16, fontSize: '0.95em' }}>
        米・調味料・乾物など、まとめて発注する食材の目安を計算します。
      </p>
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
          {loading ? '集計中...' : 'まとめ買い目安を計算'}
        </button>
      </form>

      {error && <p className="error">{error}</p>}

      {data && data.items.length > 0 && (
        <div className="card" style={{ marginTop: 16 }}>
          <h2>まとめ買い目安（{data.start} 〜 {data.end}、{data.people}人分）</h2>
          {Object.entries(byCategory(data.items)).map(([cat, items]) => (
            <div key={cat} style={{ marginBottom: 24 }}>
              <h3 style={{ fontSize: '1.1em', marginBottom: 8, color: '#333' }}>
                {CATEGORY_LABELS[cat] || cat}
              </h3>
              <table>
                <thead>
                  <tr>
                    <th>食材名</th>
                    <th>使用量</th>
                    <th>発注目安</th>
                  </tr>
                </thead>
                <tbody>
                  {items.map((item) => (
                    <tr key={item.ingredient_id}>
                      <td>{item.name}</td>
                      <td>{item.total_g.toFixed(0)}g</td>
                      <td><strong>{item.order_qty}{item.order_unit_name}</strong></td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          ))}
        </div>
      )}

      {data && data.items.length === 0 && (
        <p style={{ marginTop: 16, color: '#666' }}>
          この期間の献立では、まとめ買い対象の食材が使用されていません。
        </p>
      )}
    </div>
  )
}
