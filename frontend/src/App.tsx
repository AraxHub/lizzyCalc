import { useState } from 'react'
import './App.css'

type HistoryItem = {
  id: number
  number1: number
  number2: number
  operation: string
  result: number
  message: string
  timestamp: string
}

export default function App() {
  const [number1, setNumber1] = useState('')
  const [number2, setNumber2] = useState('')
  const [operation, setOperation] = useState('+')
  const [result, setResult] = useState<number | null>(null)
  const [error, setError] = useState('')
  const [history, setHistory] = useState<HistoryItem[]>([])
  const [loading, setLoading] = useState(false)
  const [historyLoading, setHistoryLoading] = useState(false)
  const [autoRefreshHistory, setAutoRefreshHistory] = useState(false)

  const fetchHistory = async () => {
    setHistoryLoading(true)
    try {
      const res = await fetch('/api/v1/history')
      if (!res.ok) throw new Error('Не удалось загрузить историю')
      const data = await res.json()
      setHistory(data.items || [])
    } catch (e) {
      setHistory([])
    } finally {
      setHistoryLoading(false)
    }
  }

  const handleCalculate = async (e: React.FormEvent) => {
    e.preventDefault()
    setError('')
    setResult(null)
    const a = parseFloat(number1)
    const b = parseFloat(number2)
    if (Number.isNaN(a) || Number.isNaN(b)) {
      setError('Введите числа')
      return
    }
    setLoading(true)
    try {
      const res = await fetch('/api/v1/calculate', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ number1: a, number2: b, operation }),
      })
      const data = await res.json()
      if (!res.ok) {
        setError(data.message || 'Ошибка')
        return
      }
      setResult(data.result)
      if (autoRefreshHistory) fetchHistory()
    } catch (e) {
      setError('Сервер недоступен')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="app">
      <h1>lizzyCalc</h1>
      <form onSubmit={handleCalculate} className="calculator">
        <input
          type="number"
          step="any"
          placeholder="Число 1"
          value={number1}
          onChange={(e) => setNumber1(e.target.value)}
        />
        <select value={operation} onChange={(e) => setOperation(e.target.value)}>
          <option value="+">+</option>
          <option value="-">−</option>
          <option value="*">×</option>
          <option value="/">/</option>
        </select>
        <input
          type="number"
          step="any"
          placeholder="Число 2"
          value={number2}
          onChange={(e) => setNumber2(e.target.value)}
        />
        <button type="submit" disabled={loading} className="btn-calc">
          {loading ? '…' : '='}
        </button>
      </form>
      {error && <p className="error">{error}</p>}
      {result !== null && <p className="result">= {result}</p>}
      <section className="history">
        <div className="history-header">
          <h2>История</h2>
          <div className="history-actions">
            <label className="toggle">
              <input
                type="checkbox"
                checked={autoRefreshHistory}
                onChange={(e) => setAutoRefreshHistory(e.target.checked)}
              />
              <span className="toggle-slider" />
              <span className="toggle-label">Обновлять автоматически</span>
            </label>
            <button
              type="button"
              onClick={fetchHistory}
              disabled={historyLoading}
              className="btn-history"
            >
              {historyLoading ? '…' : 'Загрузить историю'}
            </button>
          </div>
        </div>
        {history.length === 0 ? (
          <p className="history-empty">Нажмите «Загрузить историю» или включите автообновление и посчитайте</p>
        ) : (
          <ul>
            {history.map((op) => (
              <li key={op.id}>
                {op.number1} {op.operation} {op.number2} = <strong>{op.result}</strong>
              </li>
            ))}
          </ul>
        )}
      </section>
    </div>
  )
}
