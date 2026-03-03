import { useState, useEffect } from 'react'
import { getDishes, getMenuByDate, createMenu, updateMenu } from '../api'
import type { Dish } from '../api'

const CATEGORIES = ['主食', '主菜', '副菜', '汁物', 'デザート'] as const

export default function MenuInput() {
  const [date, setDate] = useState('2026-04-01')
  const [dishesByCat, setDishesByCat] = useState<Record<string, Dish[]>>({})
  const [selected, setSelected] = useState<Record<string, number>>({})
  const [hasExisting, setHasExisting] = useState(false)
  const [loading, setLoading] = useState(false)
  const [loadingMenu, setLoadingMenu] = useState(false)
  const [error, setError] = useState('')
  const [success, setSuccess] = useState('')

  useEffect(() => {
    Promise.all(CATEGORIES.map((c) => getDishes(c))).then((results) => {
      const map: Record<string, Dish[]> = {}
      CATEGORIES.forEach((c, i) => {
        map[c] = results[i]
      })
      setDishesByCat(map)
    })
  }, [])

  useEffect(() => {
    if (!date) {
      setSelected({})
      setHasExisting(false)
      setError('')
      setLoadingMenu(false)
      return
    }
    setLoadingMenu(true)
    setError('')
    setSuccess('')
    getMenuByDate(date)
      .then((menu) => {
        setSelected({
          主食: menu.staple_id ?? 0,
          主菜: menu.main_id ?? 0,
          副菜: menu.side_id ?? 0,
          汁物: menu.soup_id ?? 0,
          デザート: menu.dessert_id ?? 0,
        })
        setHasExisting(true)
      })
      .catch((e) => {
        setSelected({})
        setHasExisting(false)
        // 献立未登録（404）は正常ケースなのでエラー表示しない
        if (!e.message.includes('登録されていません')) {
          setError(toUserMessage(e.message))
        }
      })
      .finally(() => setLoadingMenu(false))
  }, [date])

  const toUserMessage = (apiError: string): string => {
    if (apiError.includes('month') || apiError.includes('date') || apiError.includes('指定')) {
      return '献立の日付を選択してください。'
    }
    if (apiError.includes('既に登録')) {
      return 'この日付の献立は既に登録されています。更新する場合は献立を変更して「献立を更新」を押してください。'
    }
    if (apiError.includes('登録されていません')) {
      return 'この日付には献立が登録されていません。新規登録の場合は「献立を登録」を押してください。'
    }
    return apiError
  }

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!date) {
      setError('献立の日付を選択してください。')
      return
    }
    const ids = CATEGORIES.map((c) => selected[c])
    if (ids.some((id) => !id)) {
      setError('主食・主菜・副菜・汁物・デザートをすべて選択してください。')
      return
    }
    setLoading(true)
    setError('')
    setSuccess('')
    const doSave = hasExisting ? updateMenu : createMenu
    doSave(date, ids[0], ids[1], ids[2], ids[3], ids[4])
      .then(() => {
        setSuccess(hasExisting ? `献立を更新しました: ${date}` : `献立を登録しました: ${date}`)
        setHasExisting(true)
        if (!hasExisting) setSelected({})
      })
      .catch((e) => setError(toUserMessage(e.message)))
      .finally(() => setLoading(false))
  }

  return (
    <div className="card">
      <h1>献立入力</h1>
      <form onSubmit={handleSubmit}>
        <div className="form-group">
          <label>日付</label>
          <input
            type="date"
            value={date}
            onChange={(e) => {
              setDate(e.target.value)
              setError('')
            }}
            required
          />
          {loadingMenu && <span style={{ marginLeft: 8, color: '#666' }}>献立を読み込み中...</span>}
          {!date && <p style={{ marginTop: 4, color: '#666', fontSize: '0.9em' }}>日付を選択すると、既存の献立があれば表示されます。</p>}
        </div>
        {CATEGORIES.map((cat) => (
          <div key={cat} className="form-group">
            <label>{cat}</label>
            <select
              value={selected[cat] ?? ''}
              onChange={(e) => setSelected({ ...selected, [cat]: Number(e.target.value) })}
              required
            >
              <option value="">選択してください</option>
              {(dishesByCat[cat] ?? []).map((d) => (
                <option key={d.id} value={d.id}>
                  {d.name}
                </option>
              ))}
            </select>
          </div>
        ))}
        {error && <p className="error" role="alert">{error}</p>}
        {success && <p style={{ color: '#16a34a' }}>{success}</p>}
        <button type="submit" disabled={loading || loadingMenu}>
          {loading ? (hasExisting ? '更新中...' : '登録中...') : hasExisting ? '献立を更新' : '献立を登録'}
        </button>
      </form>
    </div>
  )
}
