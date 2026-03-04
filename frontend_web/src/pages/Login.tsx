import { useState } from 'react'
import { useNavigate, Link } from 'react-router-dom'
import { login, register, isLoggedIn } from '../api'

export default function Login() {
  const navigate = useNavigate()
  const [mode, setMode] = useState<'login' | 'register'>('login')
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [error, setError] = useState('')

  if (isLoggedIn()) {
    navigate('/')
    return null
  }

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setError('')
    try {
      if (mode === 'login') {
        const res = await login(email, password)
        localStorage.setItem('token', res.token)
        localStorage.setItem('user', JSON.stringify(res.user))
        navigate('/')
      } else {
        await register(email, password)
        setError('')
        setMode('login')
        setPassword('')
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : 'エラーが発生しました')
    }
  }

  return (
    <div className="login-page">
      <div className="login-card">
        <h1>管理栄養士業務システム</h1>
        <h2>{mode === 'login' ? 'ログイン' : '初回登録'}</h2>
        <form onSubmit={handleSubmit}>
          <div>
            <label>メールアドレス</label>
            <input
              type="email"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              required
              autoComplete="email"
            />
          </div>
          <div>
            <label>パスワード</label>
            <input
              type="password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              required
              minLength={mode === 'register' ? 6 : undefined}
              autoComplete={mode === 'login' ? 'current-password' : 'new-password'}
            />
            {mode === 'register' && (
              <small>6文字以上</small>
            )}
          </div>
          {error && <p className="error">{error}</p>}
          <button type="submit">{mode === 'login' ? 'ログイン' : '登録'}</button>
        </form>
        <p>
          {mode === 'login' ? (
            <>初めての方は <button type="button" className="link" onClick={() => { setMode('register'); setError(''); }}>登録</button></>
          ) : (
            <>既に登録済みの方は <button type="button" className="link" onClick={() => { setMode('login'); setError(''); }}>ログイン</button></>
          )}
        </p>
        <p><Link to="/">トップに戻る</Link></p>
      </div>
    </div>
  )
}
